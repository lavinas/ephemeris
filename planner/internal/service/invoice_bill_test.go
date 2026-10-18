package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

type mockBillingClient struct {
	createdInvoices   []port.BillingInvoiceInput
	mockCreateID      int64
	mockCreateErr     error
	mockExternalInvs  map[string]int64
	failCreateOnCode  string
	cancelledInvoices []port.BillingCancelInput
	mockCancelErr     error
}

func (m *mockBillingClient) CreateInvoice(ctx context.Context, input port.BillingInvoiceInput) (int64, error) {
	m.createdInvoices = append(m.createdInvoices, input)
	if m.mockCreateErr != nil {
		return 0, m.mockCreateErr
	}
	if input.ExternalCode != nil && *input.ExternalCode == m.failCreateOnCode {
		return 0, errors.New("duplicate external_code '" + *input.ExternalCode + "' for vendor")
	}
	if m.mockCreateID > 0 {
		return m.mockCreateID, nil
	}
	return 999, nil
}

func (m *mockBillingClient) GetInvoiceByExternalCode(ctx context.Context, vendor, externalCode string) (*int64, error) {
	if m.mockExternalInvs != nil {
		if id, ok := m.mockExternalInvs[vendor+":"+externalCode]; ok {
			return &id, nil
		}
	}
	return nil, nil
}

func (m *mockBillingClient) CancelInvoice(ctx context.Context, input port.BillingCancelInput) error {
	m.cancelledInvoices = append(m.cancelledInvoices, input)
	return m.mockCancelErr
}

type mockInvoiceBillRepo struct {
	mockPlanRepoForService
	invoices       map[int64]*domain.Invoice
	updatedInvoice *domain.Invoice
}

func (m *mockInvoiceBillRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	if inv, ok := m.invoices[id]; ok {
		return inv, nil
	}
	return nil, nil
}

func (m *mockInvoiceBillRepo) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	var list []domain.Invoice
	for _, inv := range m.invoices {
		list = append(list, *inv)
	}
	return list, nil
}

func (m *mockInvoiceBillRepo) UpdateInvoice(invoice *domain.Invoice) error {
	m.updatedInvoice = invoice
	m.invoices[invoice.ID] = invoice
	return nil
}

func (m *mockInvoiceBillRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "vend_nick"}, nil
}

func (m *mockInvoiceBillRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, VendorID: 1, Nickname: "cust_nick"}, nil
}

func TestInvoiceBill_Success(t *testing.T) {
	vendor := &domain.Vendor{ID: 1, Nickname: "vendedor_teste"}
	customer := &domain.Customer{ID: 10, VendorID: 1, Nickname: "cliente_teste", Vendor: vendor}
	plan := &domain.Plan{ID: 100, CustomerID: 10, Customer: customer}
	desc := "Pilates Session"
	inv := &domain.Invoice{
		ID:            55,
		PlanID:        100,
		Plan:          plan,
		InvoicingDate: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC),
		Amount:        200.0,
		Items: []domain.InvoiceItem{
			{ID: 1, Description: &desc, Price: 100.0, InvoicedQuantity: 2},
		},
	}

	repo := &mockInvoiceBillRepo{
		invoices: map[int64]*domain.Invoice{55: inv},
	}
	billingClient := &mockBillingClient{mockCreateID: 42}
	logger := &mockLogger{}

	svc := NewInvoiceBill(repo, logger, billingClient)
	req := &dto.InvoiceBillRequest{
		InvoiceID: 55,
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	billResp := resp.(*dto.InvoiceBillResponse)
	if billResp.BilledCount != 1 {
		t.Errorf("expected 1 billed invoice, got %d", billResp.BilledCount)
	}

	if len(billingClient.createdInvoices) != 1 {
		t.Fatalf("expected 1 billing request, got %d", len(billingClient.createdInvoices))
	}

	sent := billingClient.createdInvoices[0]
	if sent.Vendor != "vendedor_teste" || sent.Customer != "cliente_teste" {
		t.Errorf("unexpected vendor/customer sent: %s / %s", sent.Vendor, sent.Customer)
	}
	if sent.Invoicing != "2026-10-15" || sent.Due != "2026-10-25" {
		t.Errorf("unexpected dates: %s / %s", sent.Invoicing, sent.Due)
	}
	if sent.ExternalCode == nil || *sent.ExternalCode != "55" {
		t.Errorf("expected external_code '55', got %v", sent.ExternalCode)
	}
	if len(sent.Items) != 1 || sent.Items[0].Description != "Pilates Session" || sent.Items[0].Price != 100.0 || sent.Items[0].Quantity != 2 {
		t.Errorf("unexpected items sent: %+v", sent.Items)
	}

	if repo.updatedInvoice == nil || repo.updatedInvoice.ReferenceInvoiceID == nil || *repo.updatedInvoice.ReferenceInvoiceID != 42 {
		t.Errorf("expected ReferenceInvoiceID 42 persisted, got %+v", repo.updatedInvoice)
	}
}

func TestInvoiceBill_AlreadyInvoiced_BadRequest(t *testing.T) {
	refID := int64(99)
	inv := &domain.Invoice{
		ID:                 60,
		ReferenceInvoiceID: &refID,
	}
	repo := &mockInvoiceBillRepo{
		invoices: map[int64]*domain.Invoice{60: inv},
	}
	billingClient := &mockBillingClient{}
	logger := &mockLogger{}

	svc := NewInvoiceBill(repo, logger, billingClient)
	req := &dto.InvoiceBillRequest{
		InvoiceID: 60,
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 400 {
		t.Fatalf("expected status 400 Bad Request, got %d", resp.GetStatusCode())
	}

	if len(billingClient.createdInvoices) > 0 {
		t.Errorf("billing client should NOT have been called for already invoiced pre-invoice")
	}
}

func TestInvoiceBill_DuplicateRecovery(t *testing.T) {
	vendor := &domain.Vendor{ID: 1, Nickname: "vendedor_teste"}
	customer := &domain.Customer{ID: 10, VendorID: 1, Nickname: "cliente_teste", Vendor: vendor}
	plan := &domain.Plan{ID: 100, CustomerID: 10, Customer: customer}
	inv := &domain.Invoice{
		ID:            70,
		PlanID:        100,
		Plan:          plan,
		InvoicingDate: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC),
		Amount:        100.0,
		Items: []domain.InvoiceItem{
			{ID: 1, Price: 100.0, InvoicedQuantity: 1},
		},
	}

	repo := &mockInvoiceBillRepo{
		invoices: map[int64]*domain.Invoice{70: inv},
	}
	billingClient := &mockBillingClient{
		failCreateOnCode: "70",
		mockExternalInvs: map[string]int64{
			"vendedor_teste:70": 303,
		},
	}
	logger := &mockLogger{}

	svc := NewInvoiceBill(repo, logger, billingClient)
	req := &dto.InvoiceBillRequest{
		InvoiceID: 70,
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200 after duplicate recovery, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if repo.updatedInvoice == nil || repo.updatedInvoice.ReferenceInvoiceID == nil || *repo.updatedInvoice.ReferenceInvoiceID != 303 {
		t.Errorf("expected ReferenceInvoiceID 303 persisted, got %+v", repo.updatedInvoice)
	}
}
