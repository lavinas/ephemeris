package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
)

type taxMockEmitter struct {
	shouldFail bool
	err        error
	nfeNumber  int64
	nfeVerif   string
	calledWith int64
}

func (m *taxMockEmitter) EmitSingle(ctx context.Context, invoice *domain.Invoice, rpsNumber int64) (*domain.EmissionItem, error) {
	m.calledWith = rpsNumber
	if m.shouldFail {
		if m.err != nil {
			return nil, m.err
		}
		return nil, errors.New("prefeitura: erro cadastral do tomador")
	}
	now := time.Now()
	amount := invoice.Amount
	nfeNum := m.nfeNumber
	if nfeNum == 0 {
		nfeNum = 12345
	}
	verif := m.nfeVerif
	if verif == "" {
		verif = "ABC123XYZ"
	}
	return &domain.EmissionItem{
		InvoiceID:       invoice.ID,
		RPSNumber:       rpsNumber,
		NFENumber:       &nfeNum,
		NFEDatetime:     &now,
		NFEVerification: &verif,
		NFEAmount:       &amount,
	}, nil
}

type taxMockRepo struct {
	*paymentMockRepo
	vendor        *domain.Vendor
	lastRPS       int64
	savedEmission *domain.Emission
	savedItem     *domain.EmissionItem
}

func newTaxMockRepo() *taxMockRepo {
	doc := "42343223858"
	email := "cliente@teste.com"
	pm := newPaymentMockRepo()
	pm.invoices[1] = &domain.Invoice{
		ID:          1,
		VendorID:    1,
		CustomerID:  1,
		Amount:      150.00,
		InvoiceDate: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		DueDate:     time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC),
		Customer: domain.Customer{
			ID:       1,
			VendorID: 1,
			Name:     "Cliente Teste",
			Nickname: "cliente_teste",
			Document: &doc,
			Email:    &email,
		},
		InvoiceItems: []domain.InvoiceItem{
			{ID: 1, InvoiceID: 1, Description: "Aula de Música", Quantity: 1, Price: 150.00},
		},
	}
	vendor := &domain.Vendor{
		ID:          1,
		Nickname:    "estudio_amelia",
		LegalName:   "Estudio Amelia LTDA",
		Document:    "27928875000104",
		TaxDocument: "57278881",
		LastRps:     2435,
	}
	return &taxMockRepo{
		paymentMockRepo: pm,
		vendor:          vendor,
		lastRPS:         2435,
	}
}

func (m *taxMockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	if m.vendor != nil && m.vendor.Nickname == nickname {
		return m.vendor, nil
	}
	return nil, fmt.Errorf("vendor '%s' not found", nickname)
}

func (m *taxMockRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	if m.vendor != nil && m.vendor.ID == id {
		return m.vendor, nil
	}
	return nil, fmt.Errorf("vendor with ID %d not found", id)
}

func (m *taxMockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return m.lastRPS, nil
}

func (m *taxMockRepo) Save(model interface{}) error {
	switch v := model.(type) {
	case *domain.Emission:
		m.savedEmission = v
	case *domain.EmissionItem:
		m.savedItem = v
	case *domain.Vendor:
		m.vendor = v
	}
	return m.paymentMockRepo.Save(model)
}

func TestInvoiceTaxEmit_Success(t *testing.T) {
	repo := newTaxMockRepo()
	logger := &mockLogger{}
	emitter := &taxMockEmitter{nfeNumber: 9988, nfeVerif: "TEST-VERIF"}

	// Mark invoice as paid
	paidDate := time.Now()
	repo.invoices[1].PaymentDate = &paidDate

	svc := NewInvoiceTaxEmit(repo, logger, emitter)
	req := &dto.InvoiceTaxEmitRequest{
		Vendor: "estudio_amelia",
		ID:     1,
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d", out.GetStatusCode())
	}

	resp, ok := out.(dto.InvoiceTaxEmitResponse)
	if !ok {
		t.Fatalf("expected InvoiceTaxEmitResponse")
	}

	if resp.NFENumber == nil || *resp.NFENumber != 9988 {
		t.Errorf("expected NFE number 9988, got %v", resp.NFENumber)
	}
	if resp.NFEVerification == nil || *resp.NFEVerification != "TEST-VERIF" {
		t.Errorf("expected verification TEST-VERIF, got %v", resp.NFEVerification)
	}
	if resp.RPSNumber != 2436 {
		t.Errorf("expected RPS number 2436, got %d", resp.RPSNumber)
	}

	// Verify persistence
	if repo.invoices[1].TaxDate == nil {
		t.Errorf("expected invoice TaxDate to be set")
	}
	if repo.vendor.LastRps != 2436 {
		t.Errorf("expected vendor LastRps to be updated to 2436, got %d", repo.vendor.LastRps)
	}
	if repo.savedEmission == nil {
		t.Errorf("expected emission to be saved")
	}
	if repo.savedItem == nil {
		t.Errorf("expected emission item to be saved")
	}
}

func TestInvoiceTaxEmit_NotPaid(t *testing.T) {
	repo := newTaxMockRepo()
	logger := &mockLogger{}
	emitter := &taxMockEmitter{}

	// Invoice has PaymentDate == nil
	svc := NewInvoiceTaxEmit(repo, logger, emitter)
	req := &dto.InvoiceTaxEmitRequest{
		Vendor: "estudio_amelia",
		ID:     1,
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 400 {
		t.Fatalf("expected status 400 for unpaid invoice, got %d", out.GetStatusCode())
	}
}

func TestInvoiceTaxEmit_AlreadyTaxed(t *testing.T) {
	repo := newTaxMockRepo()
	logger := &mockLogger{}
	emitter := &taxMockEmitter{}

	paidDate := time.Now()
	taxDate := time.Now()
	repo.invoices[1].PaymentDate = &paidDate
	repo.invoices[1].TaxDate = &taxDate

	svc := NewInvoiceTaxEmit(repo, logger, emitter)
	req := &dto.InvoiceTaxEmitRequest{
		Vendor: "estudio_amelia",
		ID:     1,
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 400 {
		t.Fatalf("expected status 400 for already taxed invoice, got %d", out.GetStatusCode())
	}
}

func TestInvoiceTaxEmit_CityHallRejection(t *testing.T) {
	repo := newTaxMockRepo()
	logger := &mockLogger{}
	emitter := &taxMockEmitter{
		shouldFail: true,
		err:        errors.New("prefeitura: [1205] CPF do tomador inválido"),
	}

	paidDate := time.Now()
	repo.invoices[1].PaymentDate = &paidDate

	svc := NewInvoiceTaxEmit(repo, logger, emitter)
	req := &dto.InvoiceTaxEmitRequest{
		Vendor: "estudio_amelia",
		ID:     1,
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 400 {
		t.Fatalf("expected status 400, got %d", out.GetStatusCode())
	}

	resp := out.(dto.InvoiceTaxEmitResponse)
	if resp.Message == "" {
		t.Errorf("expected rejection message from prefeitura")
	}

	// Verify invoice tax date remained nil
	if repo.invoices[1].TaxDate != nil {
		t.Errorf("expected invoice TaxDate to remain nil after failure")
	}
}

func TestInvoicePayment_AutomaticTaxEmission(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("could not chdir to billing root: %v", err)
	}

	repo := newTaxMockRepo()
	logger := &mockLogger{}
	issuer := &paymentMockIssuer{}
	pixer := &paymentMockPixer{}
	emitter := &taxMockEmitter{nfeNumber: 7771, nfeVerif: "AUTO-TAX-OK"}

	// Payment request
	paymentSvc := NewInvoicePayment(repo, logger, issuer, pixer, emitter)
	req := &dto.InvoicePaymentRequest{
		Vendor:      "estudio_amelia",
		ID:          1,
		PaymentDate: "2026-10-27",
	}

	out := paymentSvc.Run(req)
	if out.GetStatusCode() != 200 {
		msg := "unknown"
		if resp, ok := out.(dto.InvoicePaymentResponse); ok {
			msg = resp.Message
		}
		t.Fatalf("expected payment status 200, got %d (msg: %s)", out.GetStatusCode(), msg)
	}

	// Invoice should now have both PaymentDate and TaxDate set
	if repo.invoices[1].PaymentDate == nil {
		t.Errorf("expected PaymentDate to be set")
	}
	if repo.invoices[1].TaxDate == nil {
		t.Errorf("expected TaxDate to be automatically set by InvoicePayment")
	}
	if repo.vendor.LastRps != 2436 {
		t.Errorf("expected vendor LastRps to be incremented to 2436, got %d", repo.vendor.LastRps)
	}
}
