package dto

import (
	"strings"
	"testing"
	"time"

	"billing/internal/domain"
)

type invoiceCreateMockRepo struct {
	vendor   *domain.Vendor
	customer *domain.Customer
	invoices map[string]*domain.Invoice
}

func (m *invoiceCreateMockRepo) Save(model interface{}) error { return nil }
func (m *invoiceCreateMockRepo) BeginTransaction() error      { return nil }
func (m *invoiceCreateMockRepo) CommitTransaction() error     { return nil }
func (m *invoiceCreateMockRepo) RollbackTransaction() error   { return nil }
func (m *invoiceCreateMockRepo) Close() error                 { return nil }
func (m *invoiceCreateMockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	if m.customer != nil && m.customer.Nickname == nickname {
		return m.customer, nil
	}
	return nil, nil
}
func (m *invoiceCreateMockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	if m.vendor != nil && m.vendor.Nickname == nickname {
		return m.vendor, nil
	}
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	if m.vendor != nil && m.vendor.ID == id {
		return m.vendor, nil
	}
	return nil, nil
}
func (m *invoiceCreateMockRepo) FindInvoices(page, pageSize int, customer int64, invoiceDate, dueDate, paymentDate, sentDate, receiptDate, taxDate, cancellationDate, externalCode *string) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *invoiceCreateMockRepo) FindInvoicesPendingSend(dueBeforeOrEqual time.Time) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	if m.invoices != nil {
		for _, inv := range m.invoices {
			if inv.ID == id {
				return inv, nil
			}
		}
	}
	if m.vendor != nil {
		return &domain.Invoice{
			ID:         id,
			CustomerID: 20,
			Customer:   domain.Customer{ID: 20, VendorID: m.vendor.ID},
		}, nil
	}
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetInvoiceByExternalCode(vendorID int64, externalCode string) (*domain.Invoice, error) {
	if m.invoices != nil {
		if inv, ok := m.invoices[externalCode]; ok {
			return inv, nil
		}
	}
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}
func (m *invoiceCreateMockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}
func (m *invoiceCreateMockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

func TestInvoiceCreate_ExternalCode_Validation(t *testing.T) {
	vendor := &domain.Vendor{ID: 10, Nickname: "estudio_amelia"}
	customer := &domain.Customer{ID: 20, VendorID: 10, Nickname: "cliente_1"}
	existingCode := "EXT-100"
	repo := &invoiceCreateMockRepo{
		vendor:   vendor,
		customer: customer,
		invoices: map[string]*domain.Invoice{
			existingCode: {ID: 1, VendorID: 10, ExternalCode: &existingCode},
		},
	}

	code := "EXT-100"
	req := &InvoiceCreateRequest{
		Items: []*InvoiceCreate{
			{
				Vendor:       "estudio_amelia",
				Customer:     "cliente_1",
				InvoiceDate:  "2026-10-10",
				DueDate:      "2026-10-25",
				ExternalCode: &code,
				InvoiceItems: []*InvoiceCreateItem{
					{Description: "Item 1", Quantity: 1, Price: 100.0},
				},
			},
		},
	}

	err := req.Validate(repo)
	if err == nil {
		t.Fatalf("expected error for duplicate external_code, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate external_code 'EXT-100' for vendor") {
		t.Errorf("unexpected error message: %v", err)
	}

	// Unique code succeeds
	uniqueCode := "EXT-200"
	req.Items[0].ExternalCode = &uniqueCode
	if err := req.Validate(repo); err != nil {
		t.Fatalf("unexpected error for unique external_code: %v", err)
	}

	// Verify domain invoice contains vendor_id and external_code
	domainInvs, err := req.GetDomain()
	if err != nil {
		t.Fatalf("unexpected error from GetDomain: %v", err)
	}
	if len(domainInvs) != 1 {
		t.Fatalf("expected 1 domain invoice, got %d", len(domainInvs))
	}
	if domainInvs[0].VendorID != 10 {
		t.Errorf("expected VendorID 10, got %d", domainInvs[0].VendorID)
	}
	if domainInvs[0].ExternalCode == nil || *domainInvs[0].ExternalCode != "EXT-200" {
		t.Errorf("expected ExternalCode EXT-200, got %v", domainInvs[0].ExternalCode)
	}
}

func TestInvoiceCreate_InBatch_DuplicateExternalCode(t *testing.T) {
	vendor := &domain.Vendor{ID: 10, Nickname: "estudio_amelia"}
	customer := &domain.Customer{ID: 20, VendorID: 10, Nickname: "cliente_1"}
	repo := &invoiceCreateMockRepo{
		vendor:   vendor,
		customer: customer,
	}

	code1 := "DUPLICATE-BATCH"
	code2 := "DUPLICATE-BATCH"
	req := &InvoiceCreateRequest{
		Items: []*InvoiceCreate{
			{
				Vendor:       "estudio_amelia",
				Customer:     "cliente_1",
				InvoiceDate:  "2026-10-10",
				DueDate:      "2026-10-25",
				ExternalCode: &code1,
				InvoiceItems: []*InvoiceCreateItem{
					{Description: "Item 1", Quantity: 1, Price: 100.0},
				},
			},
			{
				Vendor:       "estudio_amelia",
				Customer:     "cliente_1",
				InvoiceDate:  "2026-10-10",
				DueDate:      "2026-10-25",
				ExternalCode: &code2,
				InvoiceItems: []*InvoiceCreateItem{
					{Description: "Item 2", Quantity: 1, Price: 200.0},
				},
			},
		},
	}

	err := req.Validate(repo)
	if err == nil {
		t.Fatalf("expected duplicate external_code in batch error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate external_code 'DUPLICATE-BATCH' in request") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestInvoiceCreateResponse_IDs(t *testing.T) {
	resp := NewInvoiceCreateResponse(200, "success", "Invoice created successfully", 42, 43)
	if resp.ID != 42 {
		t.Errorf("expected ID to be 42, got %d", resp.ID)
	}
	if len(resp.IDs) != 2 || resp.IDs[0] != 42 || resp.IDs[1] != 43 {
		t.Errorf("expected IDs to be [42, 43], got %v", resp.IDs)
	}
}
