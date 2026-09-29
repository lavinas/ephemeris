package service

import (
	"os"
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

type autoBillMockRepo struct {
	invoices map[int64]*domain.Invoice
	vendors  map[int64]*domain.Vendor
}

func newAutoBillMockRepo() *autoBillMockRepo {
	return &autoBillMockRepo{
		invoices: make(map[int64]*domain.Invoice),
		vendors:  make(map[int64]*domain.Vendor),
	}
}

func (m *autoBillMockRepo) Save(model interface{}) error {
	if inv, ok := model.(*domain.Invoice); ok {
		copied := *inv
		m.invoices[inv.ID] = &copied
	}
	return nil
}

func (m *autoBillMockRepo) BeginTransaction() error    { return nil }
func (m *autoBillMockRepo) CommitTransaction() error   { return nil }
func (m *autoBillMockRepo) RollbackTransaction() error { return nil }
func (m *autoBillMockRepo) Close() error               { return nil }

func (m *autoBillMockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}

func (m *autoBillMockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}

func (m *autoBillMockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}

func (m *autoBillMockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	for _, v := range m.vendors {
		if v.Nickname == nickname {
			copied := *v
			return &copied, nil
		}
	}
	return nil, nil
}

func (m *autoBillMockRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	if v, ok := m.vendors[id]; ok {
		copied := *v
		return &copied, nil
	}
	return nil, nil
}

func (m *autoBillMockRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, emailReceiptDate, whatsappReceiptDate,
	taxDate, cancellationDate *string) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *autoBillMockRepo) FindInvoicesPendingSend(dueBeforeOrEqual time.Time) ([]domain.Invoice, error) {
	var list []domain.Invoice
	for _, inv := range m.invoices {
		if (inv.DueDate.Before(dueBeforeOrEqual) || inv.DueDate.Equal(dueBeforeOrEqual)) &&
			inv.EmailSentDate == nil && inv.CancellationDate == nil {
			copied := *inv
			list = append(list, copied)
		}
	}
	return list, nil
}

func (m *autoBillMockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *autoBillMockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	if inv, ok := m.invoices[id]; ok {
		copied := *inv
		return &copied, nil
	}
	return nil, nil
}

func (m *autoBillMockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}

func (m *autoBillMockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}

func (m *autoBillMockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

type autoBillMockIssuer struct {
	sentCount int
}

func (i *autoBillMockIssuer) GetBase64(data port.InDTO, html_pdf string) ([]byte, error) {
	return []byte("pdf"), nil
}

func (i *autoBillMockIssuer) SendMail(data port.InDTO, subject, filename, html_pdf, html_email string) error {
	i.sentCount++
	return nil
}

func TestInvoiceAutoBill_Success(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("could not chdir to billing root: %v", err)
	}

	repo := newAutoBillMockRepo()
	repo.vendors[1] = &domain.Vendor{
		ID:          1,
		Nickname:    "estudio_amelia",
		TradingName: "Estúdio Amélia Cardoso",
		Email:       "amelia@example.com",
	}

	email1 := "cliente1@example.com"
	email2 := "cliente2@example.com"
	now := time.Now()

	// Invoice 1: Due in 2 days, not sent -> should be sent
	repo.invoices[1] = &domain.Invoice{
		ID:          1,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Name: "Cliente Um", Email: &email1},
		Amount:      100.0,
		InvoiceDate: now.AddDate(0, 0, -10),
		DueDate:     now.AddDate(0, 0, 2),
		InvoiceItems: []domain.InvoiceItem{
			{ID: 1, Description: "Item 1", Quantity: 1, Price: 100.0},
		},
	}

	// Invoice 2: Due in 1 day, not sent -> should be sent
	repo.invoices[2] = &domain.Invoice{
		ID:          2,
		CustomerID:  2,
		Customer:    domain.Customer{ID: 2, VendorID: 1, Nickname: "cliente_2", Name: "Cliente Dois", Email: &email2},
		Amount:      200.0,
		InvoiceDate: now.AddDate(0, 0, -10),
		DueDate:     now.AddDate(0, 0, 1),
		InvoiceItems: []domain.InvoiceItem{
			{ID: 2, Description: "Item 2", Quantity: 1, Price: 200.0},
		},
	}

	// Invoice 3: Due in 10 days -> should NOT be sent
	repo.invoices[3] = &domain.Invoice{
		ID:          3,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Name: "Cliente Um", Email: &email1},
		Amount:      300.0,
		InvoiceDate: now.AddDate(0, 0, -5),
		DueDate:     now.AddDate(0, 0, 10),
		InvoiceItems: []domain.InvoiceItem{
			{ID: 3, Description: "Item 3", Quantity: 1, Price: 300.0},
		},
	}

	// Invoice 4: Due in 1 day, but already sent -> should NOT be sent
	sentDate := now.AddDate(0, 0, -2)
	repo.invoices[4] = &domain.Invoice{
		ID:            4,
		CustomerID:    1,
		Customer:      domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_1", Name: "Cliente Um", Email: &email1},
		Amount:        400.0,
		InvoiceDate:   now.AddDate(0, 0, -10),
		DueDate:       now.AddDate(0, 0, 1),
		EmailSentDate: &sentDate,
		InvoiceItems: []domain.InvoiceItem{
			{ID: 4, Description: "Item 4", Quantity: 1, Price: 400.0},
		},
	}

	logger := &mockLogger{}
	issuer := &autoBillMockIssuer{}
	pixer := &paymentMockPixer{}
	billSvc := NewBill(repo, logger, issuer, pixer)

	autoBillSvc := NewInvoiceAutoBill(repo, logger, billSvc, 3)

	res := autoBillSvc.Run(nil)
	resp, ok := res.(*dto.InvoiceAutoBillResponse)
	if !ok {
		t.Fatalf("expected *dto.InvoiceAutoBillResponse, got %T", res)
	}

	if resp.HttpCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.HttpCode, resp.Message)
	}

	if resp.TotalFound != 2 {
		t.Errorf("expected 2 invoices found, got %d", resp.TotalFound)
	}
	if resp.TotalSent != 2 {
		t.Errorf("expected 2 invoices sent, got %d", resp.TotalSent)
	}
	if resp.TotalErrors != 0 {
		t.Errorf("expected 0 errors, got %d", resp.TotalErrors)
	}
	if issuer.sentCount != 2 {
		t.Errorf("expected issuer to send 2 emails, got %d", issuer.sentCount)
	}

	// Verify EmailSentDate was marked in repository
	if repo.invoices[1].EmailSentDate == nil {
		t.Errorf("expected invoice 1 EmailSentDate to be populated")
	}
	if repo.invoices[2].EmailSentDate == nil {
		t.Errorf("expected invoice 2 EmailSentDate to be populated")
	}

	// Subsequent execution should find 0 pending invoices
	res2 := autoBillSvc.Run(nil)
	resp2 := res2.(*dto.InvoiceAutoBillResponse)
	if resp2.TotalFound != 0 {
		t.Errorf("expected 0 invoices found on second run, got %d", resp2.TotalFound)
	}
}
