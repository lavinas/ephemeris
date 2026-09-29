package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

type paymentMockRepo struct {
	invoices      map[int64]*domain.Invoice
	txInProgress  bool
	committed     bool
	rolledBack    bool
	savedInvoices []*domain.Invoice
}

func newPaymentMockRepo() *paymentMockRepo {
	return &paymentMockRepo{
		invoices: make(map[int64]*domain.Invoice),
	}
}

func (m *paymentMockRepo) Save(model interface{}) error {
	if inv, ok := model.(*domain.Invoice); ok {
		copied := *inv
		m.invoices[inv.ID] = &copied
		m.savedInvoices = append(m.savedInvoices, &copied)
	}
	return nil
}

func (m *paymentMockRepo) BeginTransaction() error {
	if m.txInProgress {
		return errors.New("tx already in progress")
	}
	m.txInProgress = true
	m.committed = false
	m.rolledBack = false
	return nil
}

func (m *paymentMockRepo) CommitTransaction() error {
	if !m.txInProgress {
		return errors.New("no tx in progress")
	}
	m.txInProgress = false
	m.committed = true
	return nil
}

func (m *paymentMockRepo) RollbackTransaction() error {
	if !m.txInProgress {
		return errors.New("no tx in progress")
	}
	m.txInProgress = false
	m.rolledBack = true
	return nil
}

func (m *paymentMockRepo) Close() error { return nil }

func (m *paymentMockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}

func (m *paymentMockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}

func (m *paymentMockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}

func (m *paymentMockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	if nickname == "estudio_amelia" {
		return &domain.Vendor{
			ID:          1,
			Nickname:    "estudio_amelia",
			TradingName: "Estudio Amelia",
			Email:       "amelia@example.com",
		}, nil
	}
	return nil, nil
}

func (m *paymentMockRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, emailReceiptDate, whatsappReceiptDate,
	taxDate, cancellationDate *string) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *paymentMockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *paymentMockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	inv, ok := m.invoices[id]
	if !ok {
		return nil, nil
	}
	copied := *inv
	return &copied, nil
}

func (m *paymentMockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}

func (m *paymentMockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}

func (m *paymentMockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

type paymentMockIssuer struct {
	shouldFail bool
	mailSent   bool
}

func (i *paymentMockIssuer) GetBase64(data port.InDTO, html_pdf string) ([]byte, error) {
	if i.shouldFail {
		return nil, errors.New("pdf generation failed")
	}
	return []byte("pdf"), nil
}

func (i *paymentMockIssuer) SendMail(data port.InDTO, subject, filename, html_pdf, html_email string) error {
	if i.shouldFail {
		return errors.New("sendmail failed")
	}
	i.mailSent = true
	return nil
}

type paymentMockPixer struct{}

func (p *paymentMockPixer) Get(request port.InDTO) (string, string, error) {
	return "pix_code", "pix_base64", nil
}

func TestInvoicePayment_Success(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("could not chdir to billing root: %v", err)
	}

	repo := newPaymentMockRepo()
	email := "cliente@example.com"
	now := time.Now()
	repo.invoices[1] = &domain.Invoice{
		ID:          1,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Name: "Cliente Um", Email: &email},
		Amount:      150.0,
		InvoiceDate: now.AddDate(0, 0, -10),
		DueDate:     now.AddDate(0, 0, 5),
		PaymentDate: nil,
		InvoiceItems: []domain.InvoiceItem{
			{ID: 1, Description: "Aula de Música", Quantity: 1, Price: 150.0},
		},
	}

	logger := &mockLogger{}
	issuer := &paymentMockIssuer{shouldFail: false}
	pixer := &paymentMockPixer{}

	svc := NewInvoicePayment(repo, logger, issuer, pixer)

	req := &dto.InvoicePaymentRequest{
		Vendor:      "estudio_amelia",
		ID:          1,
		PaymentDate: "2026-09-29",
	}

	res := svc.Run(req)
	resp, ok := res.(dto.InvoicePaymentResponse)
	if !ok {
		t.Fatalf("expected InvoicePaymentResponse, got %T", res)
	}

	if resp.HttpCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.HttpCode, resp.Message)
	}

	if !issuer.mailSent {
		t.Errorf("expected receipt email to be sent")
	}

	if !repo.committed {
		t.Errorf("expected transaction to be committed")
	}

	updatedInv := repo.invoices[1]
	if updatedInv.PaymentDate == nil || updatedInv.PaymentDate.Format("2006-01-02") != "2026-09-29" {
		t.Errorf("expected PaymentDate to be 2026-09-29, got %v", updatedInv.PaymentDate)
	}
	if updatedInv.EmailReceiptDate == nil {
		t.Errorf("expected EmailReceiptDate to be updated")
	}
}

func TestInvoicePayment_EmailFailureRollsBack(t *testing.T) {
	repo := newPaymentMockRepo()
	email := "cliente@example.com"
	now := time.Now()
	repo.invoices[1] = &domain.Invoice{
		ID:          1,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Name: "Cliente Um", Email: &email},
		Amount:      150.0,
		InvoiceDate: now.AddDate(0, 0, -10),
		DueDate:     now.AddDate(0, 0, 5),
		PaymentDate: nil,
		InvoiceItems: []domain.InvoiceItem{
			{ID: 1, Description: "Aula de Música", Quantity: 1, Price: 150.0},
		},
	}

	logger := &mockLogger{}
	issuer := &paymentMockIssuer{shouldFail: true}
	pixer := &paymentMockPixer{}

	svc := NewInvoicePayment(repo, logger, issuer, pixer)

	req := &dto.InvoicePaymentRequest{
		Vendor:      "estudio_amelia",
		ID:          1,
		PaymentDate: "2026-09-29",
	}

	res := svc.Run(req)
	resp, ok := res.(dto.InvoicePaymentResponse)
	if !ok {
		t.Fatalf("expected InvoicePaymentResponse, got %T", res)
	}

	if resp.HttpCode != 500 {
		t.Fatalf("expected 500, got %d", resp.HttpCode)
	}

	if !repo.rolledBack {
		t.Errorf("expected transaction to be rolled back on email failure")
	}
}

func TestInvoicePayment_ValidationErrors(t *testing.T) {
	repo := newPaymentMockRepo()
	email := "cliente@example.com"
	now := time.Now()
	repo.invoices[1] = &domain.Invoice{
		ID:          1,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Email: &email},
		PaymentDate: &now, // already paid
	}
	cancelled := now
	repo.invoices[2] = &domain.Invoice{
		ID:               2,
		CustomerID:       1,
		Customer:         domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Email: &email},
		CancellationDate: &cancelled, // cancelled
	}
	repo.invoices[3] = &domain.Invoice{
		ID:          3,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Email: nil}, // no email
		PaymentDate: nil,
	}

	logger := &mockLogger{}
	issuer := &paymentMockIssuer{shouldFail: false}
	pixer := &paymentMockPixer{}

	svc := NewInvoicePayment(repo, logger, issuer, pixer)

	t.Run("Already paid error", func(t *testing.T) {
		req := &dto.InvoicePaymentRequest{
			Vendor:      "estudio_amelia",
			ID:          1,
			PaymentDate: "2026-09-29",
		}
		resp := svc.Run(req).(dto.InvoicePaymentResponse)
		if resp.HttpCode != 400 {
			t.Errorf("expected 400, got %d", resp.HttpCode)
		}
	})

	t.Run("Canceled error", func(t *testing.T) {
		req := &dto.InvoicePaymentRequest{
			Vendor:      "estudio_amelia",
			ID:          2,
			PaymentDate: "2026-09-29",
		}
		resp := svc.Run(req).(dto.InvoicePaymentResponse)
		if resp.HttpCode != 400 {
			t.Errorf("expected 400, got %d", resp.HttpCode)
		}
	})

	t.Run("No email error", func(t *testing.T) {
		req := &dto.InvoicePaymentRequest{
			Vendor:      "estudio_amelia",
			ID:          3,
			PaymentDate: "2026-09-29",
		}
		resp := svc.Run(req).(dto.InvoicePaymentResponse)
		if resp.HttpCode != 400 {
			t.Errorf("expected 400, got %d", resp.HttpCode)
		}
	})
}
