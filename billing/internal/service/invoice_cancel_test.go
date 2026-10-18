package service

import (
	"errors"
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
)

type cancelMockRepo struct {
	vendor    *domain.Vendor
	invoice   *domain.Invoice
	saved     *domain.Invoice
	saveError error
}

func (m *cancelMockRepo) Save(model interface{}) error {
	if m.saveError != nil {
		return m.saveError
	}
	if inv, ok := model.(*domain.Invoice); ok {
		m.saved = inv
	}
	return nil
}
func (m *cancelMockRepo) BeginTransaction() error    { return nil }
func (m *cancelMockRepo) CommitTransaction() error   { return nil }
func (m *cancelMockRepo) RollbackTransaction() error { return nil }
func (m *cancelMockRepo) Close() error               { return nil }

func (m *cancelMockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *cancelMockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}
func (m *cancelMockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *cancelMockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	if m.vendor != nil && m.vendor.Nickname == nickname {
		return m.vendor, nil
	}
	return nil, nil
}
func (m *cancelMockRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	if m.vendor != nil && m.vendor.ID == id {
		return m.vendor, nil
	}
	return nil, nil
}
func (m *cancelMockRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, sentDate, receiptDate,
	taxDate, cancellationDate, externalCode *string) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *cancelMockRepo) FindInvoicesPendingSend(dueBeforeOrEqual time.Time) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *cancelMockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *cancelMockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	if m.invoice != nil && m.invoice.ID == id {
		return m.invoice, nil
	}
	return nil, nil
}
func (m *cancelMockRepo) GetInvoiceByExternalCode(vendorID int64, externalCode string) (*domain.Invoice, error) {
	if m.invoice != nil && m.invoice.ExternalCode != nil && *m.invoice.ExternalCode == externalCode {
		return m.invoice, nil
	}
	return nil, nil
}
func (m *cancelMockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}
func (m *cancelMockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}
func (m *cancelMockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

func TestInvoiceCancel_Run(t *testing.T) {
	logger := &mockLogger{}

	t.Run("successful cancellation by ID", func(t *testing.T) {
		vendor := &domain.Vendor{ID: 1, Nickname: "estudio_amelia"}
		inv := &domain.Invoice{
			ID:       10,
			VendorID: 1,
			Customer: domain.Customer{ID: 2, VendorID: 1},
		}
		repo := &cancelMockRepo{vendor: vendor, invoice: inv}

		svc := NewInvoiceCancel(repo, logger)
		req := &dto.InvoiceCancelRequest{
			Vendor: "estudio_amelia",
			ID:     10,
		}

		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Fatalf("expected status 200, got %d", resp.GetStatusCode())
		}
		if repo.saved == nil || repo.saved.CancellationDate == nil {
			t.Fatalf("expected invoice to be saved with CancellationDate")
		}
	})

	t.Run("successful cancellation by ExternalCode", func(t *testing.T) {
		code := "EXT-99"
		vendor := &domain.Vendor{ID: 1, Nickname: "estudio_amelia"}
		inv := &domain.Invoice{
			ID:           20,
			VendorID:     1,
			ExternalCode: &code,
			Customer:     domain.Customer{ID: 2, VendorID: 1},
		}
		repo := &cancelMockRepo{vendor: vendor, invoice: inv}

		svc := NewInvoiceCancel(repo, logger)
		req := &dto.InvoiceCancelRequest{
			Vendor:       "estudio_amelia",
			ExternalCode: &code,
		}

		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Fatalf("expected status 200, got %d", resp.GetStatusCode())
		}
		if repo.saved == nil || repo.saved.CancellationDate == nil {
			t.Fatalf("expected invoice to be saved with CancellationDate")
		}
	})

	t.Run("validation failure returns 400", func(t *testing.T) {
		repo := &cancelMockRepo{}
		svc := NewInvoiceCancel(repo, logger)
		req := &dto.InvoiceCancelRequest{}

		resp := svc.Run(req)
		if resp.GetStatusCode() != 400 {
			t.Fatalf("expected status 400, got %d", resp.GetStatusCode())
		}
	})

	t.Run("repo save failure returns 500", func(t *testing.T) {
		vendor := &domain.Vendor{ID: 1, Nickname: "estudio_amelia"}
		inv := &domain.Invoice{
			ID:       30,
			VendorID: 1,
			Customer: domain.Customer{ID: 2, VendorID: 1},
		}
		repo := &cancelMockRepo{
			vendor:    vendor,
			invoice:   inv,
			saveError: errors.New("db connection lost"),
		}

		svc := NewInvoiceCancel(repo, logger)
		req := &dto.InvoiceCancelRequest{
			Vendor: "estudio_amelia",
			ID:     30,
		}

		resp := svc.Run(req)
		if resp.GetStatusCode() != 500 {
			t.Fatalf("expected status 500, got %d", resp.GetStatusCode())
		}
	})
}
