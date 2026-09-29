package service

import (
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
)

type mockLogger struct{}

func (m *mockLogger) IPrintf(level int, format string, v ...interface{}) {}
func (m *mockLogger) Close()                                             {}

type mockRepo struct {
	invoices []domain.Invoice
}

func (m *mockRepo) Save(model interface{}) error   { return nil }
func (m *mockRepo) BeginTransaction() error        { return nil }
func (m *mockRepo) CommitTransaction() error       { return nil }
func (m *mockRepo) RollbackTransaction() error     { return nil }
func (m *mockRepo) Close() error                   { return nil }

func (m *mockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}

func (m *mockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return &domain.Customer{ID: 1, VendorID: 1, Nickname: nickname}, nil
}

func (m *mockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}

func (m *mockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	return &domain.Vendor{ID: 1, Nickname: nickname}, nil
}

func (m *mockRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "vendor_mock"}, nil
}

func (m *mockRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, emailReceiptDate, whatsappReceiptDate,
	taxDate, cancellationDate *string) ([]domain.Invoice, error) {
	return m.invoices, nil
}

func (m *mockRepo) FindInvoicesPendingSend(dueBeforeOrEqual time.Time) ([]domain.Invoice, error) {
	return m.invoices, nil
}

func (m *mockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *mockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	return nil, nil
}

func (m *mockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}

func (m *mockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}

func (m *mockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

func TestInvoiceList_Run_OverdueFiltering(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)
	paid := today

	repo := &mockRepo{
		invoices: []domain.Invoice{
			{
				ID:          1,
				CustomerID:  1,
				Customer:    domain.Customer{ID: 1, Nickname: "cliente_1"},
				Amount:      100.0,
				InvoiceDate: yesterday,
				DueDate:     yesterday,
				PaymentDate: nil, // overdue = true
			},
			{
				ID:          2,
				CustomerID:  1,
				Customer:    domain.Customer{ID: 1, Nickname: "cliente_1"},
				Amount:      200.0,
				InvoiceDate: today,
				DueDate:     tomorrow,
				PaymentDate: nil, // overdue = false
			},
			{
				ID:          3,
				CustomerID:  1,
				Customer:    domain.Customer{ID: 1, Nickname: "cliente_1"},
				Amount:      300.0,
				InvoiceDate: yesterday,
				DueDate:     yesterday,
				PaymentDate: &paid, // overdue = false (paid)
			},
		},
	}

	logger := &mockLogger{}
	svc := NewInvoiceList(repo, logger)

	vendorStr := "estudio_amelia"
	trueVal := true
	falseVal := false

	t.Run("No overdue filter returns all invoices with overdue computed", func(t *testing.T) {
		req := &dto.InvoiceListRequest{
			Page:     1,
			PageSize: 10,
			Vendor:   &vendorStr,
		}
		resp := svc.Run(req).(dto.InvoiceListResponse)
		if resp.HttpCode != 200 {
			t.Fatalf("expected 200, got %d", resp.HttpCode)
		}
		if resp.TotalCount != 3 {
			t.Errorf("expected TotalCount = 3, got %d", resp.TotalCount)
		}
		if resp.TotalAmount != 600.0 {
			t.Errorf("expected TotalAmount = 600.0, got %f", resp.TotalAmount)
		}
		if len(resp.Invoices) != 3 {
			t.Fatalf("expected 3 invoices, got %d", len(resp.Invoices))
		}
		if !resp.Invoices[0].Overdue {
			t.Errorf("expected invoice 1 overdue = true")
		}
		if resp.Invoices[1].Overdue {
			t.Errorf("expected invoice 2 overdue = false")
		}
		if resp.Invoices[2].Overdue {
			t.Errorf("expected invoice 3 overdue = false")
		}
	})

	t.Run("Filter overdue = true", func(t *testing.T) {
		req := &dto.InvoiceListRequest{
			Page:     1,
			PageSize: 10,
			Vendor:   &vendorStr,
			Overdue:  &trueVal,
		}
		resp := svc.Run(req).(dto.InvoiceListResponse)
		if resp.HttpCode != 200 {
			t.Fatalf("expected 200, got %d", resp.HttpCode)
		}
		if resp.TotalCount != 1 {
			t.Errorf("expected TotalCount = 1, got %d", resp.TotalCount)
		}
		if resp.TotalAmount != 100.0 {
			t.Errorf("expected TotalAmount = 100.0, got %f", resp.TotalAmount)
		}
		if len(resp.Invoices) != 1 {
			t.Fatalf("expected 1 invoice, got %d", len(resp.Invoices))
		}
		if resp.Invoices[0].ID != 1 || !resp.Invoices[0].Overdue {
			t.Errorf("expected invoice ID 1 overdue = true")
		}
	})

	t.Run("Filter overdue = false", func(t *testing.T) {
		req := &dto.InvoiceListRequest{
			Page:     1,
			PageSize: 10,
			Vendor:   &vendorStr,
			Overdue:  &falseVal,
		}
		resp := svc.Run(req).(dto.InvoiceListResponse)
		if resp.HttpCode != 200 {
			t.Fatalf("expected 200, got %d", resp.HttpCode)
		}
		if resp.TotalCount != 2 {
			t.Errorf("expected TotalCount = 2, got %d", resp.TotalCount)
		}
		if resp.TotalAmount != 500.0 {
			t.Errorf("expected TotalAmount = 500.0, got %f", resp.TotalAmount)
		}
		if len(resp.Invoices) != 2 {
			t.Fatalf("expected 2 invoices, got %d", len(resp.Invoices))
		}
		for _, inv := range resp.Invoices {
			if inv.Overdue {
				t.Errorf("expected all returned invoices to have Overdue = false, got ID %d with Overdue = true", inv.ID)
			}
		}
	})
}
