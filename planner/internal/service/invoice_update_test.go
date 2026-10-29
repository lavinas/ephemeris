package service

import (
	"errors"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoiceUpdateRepoForService struct {
	mockPlanRepoForService
	invoices       map[int64]*domain.Invoice
	updatedInvoice *domain.Invoice
	failGet        bool
	failUpdate     bool
}

func newMockInvoiceUpdateRepoForService() *mockInvoiceUpdateRepoForService {
	return &mockInvoiceUpdateRepoForService{
		invoices: make(map[int64]*domain.Invoice),
	}
}

func (m *mockInvoiceUpdateRepoForService) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	if m.failGet {
		return nil, errors.New("not found")
	}
	if inv, ok := m.invoices[id]; ok && inv.DeletedAt == nil {
		return inv, nil
	}
	return nil, nil
}

func (m *mockInvoiceUpdateRepoForService) UpdateInvoice(invoice *domain.Invoice) error {
	if m.failUpdate {
		return errors.New("not found")
	}
	m.updatedInvoice = invoice
	m.invoices[invoice.ID] = invoice
	return nil
}

func (m *mockInvoiceUpdateRepoForService) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "vendedor_teste"}, nil
}

func (m *mockInvoiceUpdateRepoForService) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, VendorID: 1, Nickname: "cliente_teste"}, nil
}

func TestInvoiceUpdate_Success(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	origDueDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	desc1 := "Consulta"
	desc2 := "Exame"
	inv := &domain.Invoice{
		ID:            1,
		DueDate:       origDueDate,
		InvoicingDate: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Amount:        250.0,
		Items: []domain.InvoiceItem{
			{ID: 10, InvoiceID: 1, ServiceID: 100, Description: &desc1, Price: 100.0, InvoicedQuantity: 1, RealizedQuantity: 1},
			{ID: 20, InvoiceID: 1, ServiceID: 200, Description: &desc2, Price: 75.0, InvoicedQuantity: 2, RealizedQuantity: 2},
		},
	}
	repo.invoices[1] = inv

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 1,
		DueDate:   "2026-11-05",
		Items: []dto.InvoiceItemUpdateRequest{
			{ID: 10, InvoicedQuantity: 3}, // 3 * 100 = 300
			{ID: 20, InvoicedQuantity: 1}, // 1 * 75 = 75
		},
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d: %s", res.GetStatusCode(), res.GetMessage())
	}

	updateResp, ok := res.(*dto.InvoiceUpdateResponse)
	if !ok || updateResp.Invoice == nil {
		t.Fatalf("expected valid InvoiceUpdateResponse with InvoiceDTO")
	}

	// Verify total amount recalculated: 300 + 75 = 375
	if updateResp.Invoice.Amount != 375.0 {
		t.Errorf("expected recalculated amount 375.0, got %.2f", updateResp.Invoice.Amount)
	}

	// Verify DueDate updated
	if updateResp.Invoice.DueDate != "2026-11-05" {
		t.Errorf("expected due date 2026-11-05, got %s", updateResp.Invoice.DueDate)
	}

	// Verify repo received the updated model
	if repo.updatedInvoice == nil {
		t.Fatalf("repo did not receive updated invoice")
	}
	if repo.updatedInvoice.Amount != 375.0 {
		t.Errorf("repo invoice amount expected 375.0, got %.2f", repo.updatedInvoice.Amount)
	}
	if repo.updatedInvoice.DueDate.Format("2006-01-02") != "2026-11-05" {
		t.Errorf("repo invoice due date expected 2026-11-05, got %s", repo.updatedInvoice.DueDate.Format("2006-01-02"))
	}
	if repo.updatedInvoice.Items[0].InvoicedQuantity != 3 {
		t.Errorf("item 10 quantity expected 3, got %d", repo.updatedInvoice.Items[0].InvoicedQuantity)
	}
	if repo.updatedInvoice.Items[1].InvoicedQuantity != 1 {
		t.Errorf("item 20 quantity expected 1, got %d", repo.updatedInvoice.Items[1].InvoicedQuantity)
	}
}

func TestInvoiceUpdate_ValidationFailure(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 0,
		DueDate:   "invalid-date",
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 {
		t.Errorf("expected 400 on validation failure, got %d", res.GetStatusCode())
	}
}

func TestInvoiceUpdate_NotFound(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 999,
		DueDate:   "2026-11-05",
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 { // Validate checks repo and returns error
		t.Errorf("expected 400 when invoice not found in validation, got %d", res.GetStatusCode())
	}
}

func TestInvoiceUpdate_AlreadyInvoiced(t *testing.T) {
	t.Run("successfully cancels previous invoice and creates replacement in billing", func(t *testing.T) {
		repo := newMockInvoiceUpdateRepoForService()
		logger := &mockLogger{}
		billingClient := &mockBillingClient{
			mockCreateID: 9999,
		}

		billingID := int64(789)
		vendor := &domain.Vendor{ID: 1, Nickname: "vendedor_teste"}
		customer := &domain.Customer{ID: 10, VendorID: 1, Nickname: "cliente_teste", Vendor: vendor}
		plan := &domain.Plan{ID: 1, CustomerID: 10, Customer: customer}

		desc1 := "Consulta"
		repo.invoices[1] = &domain.Invoice{
			ID:                 1,
			DueDate:            time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
			InvoicingDate:      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			Amount:             250.0,
			ReferenceInvoiceID: &billingID,
			PlanID:             1,
			Plan:               plan,
			Items: []domain.InvoiceItem{
				{ID: 10, InvoiceID: 1, ServiceID: 100, Description: &desc1, Price: 100.0, InvoicedQuantity: 1, RealizedQuantity: 1},
			},
		}

		svc := NewInvoiceUpdate(repo, logger, billingClient)
		req := &dto.InvoiceUpdateRequest{
			InvoiceID: 1,
			DueDate:   "2026-11-05",
			Items: []dto.InvoiceItemUpdateRequest{
				{ID: 10, InvoicedQuantity: 2},
			},
		}

		res := svc.Run(req)
		if res.GetStatusCode() != 200 {
			t.Fatalf("expected 200, got %d: %s", res.GetStatusCode(), res.GetMessage())
		}

		// 1. Check old invoice cancelled in billing
		if len(billingClient.cancelledInvoices) != 1 {
			t.Fatalf("expected 1 cancelled invoice in billing, got %d", len(billingClient.cancelledInvoices))
		}
		cancelled := billingClient.cancelledInvoices[0]
		if cancelled.ID != 789 {
			t.Errorf("expected cancelled ID 789, got %d", cancelled.ID)
		}
		if cancelled.Vendor != "vendedor_teste" {
			t.Errorf("expected cancelled vendor 'vendedor_teste', got '%s'", cancelled.Vendor)
		}
		todayStr := time.Now().Format("2006-01-02")
		if cancelled.CancellationDate == nil || *cancelled.CancellationDate != todayStr {
			t.Errorf("expected cancellation date %s, got %v", todayStr, cancelled.CancellationDate)
		}

		// 2. Check replacement invoice created in billing
		if len(billingClient.createdInvoices) != 1 {
			t.Fatalf("expected 1 created invoice in billing, got %d", len(billingClient.createdInvoices))
		}
		created := billingClient.createdInvoices[0]
		if created.Vendor != "vendedor_teste" {
			t.Errorf("expected created vendor 'vendedor_teste', got '%s'", created.Vendor)
		}
		if created.Customer != "cliente_teste" {
			t.Errorf("expected created customer 'cliente_teste', got '%s'", created.Customer)
		}
		if created.Due != "2026-11-05" {
			t.Errorf("expected due date '2026-11-05', got '%s'", created.Due)
		}
		if len(created.Items) != 1 || created.Items[0].Quantity != 2 {
			t.Errorf("expected 1 item with quantity 2, got %v", created.Items)
		}

		// 3. Check pre-invoice ReferenceInvoiceID updated to new ID in repo
		if repo.updatedInvoice == nil {
			t.Fatalf("expected repo updated invoice not to be nil")
		}
		if repo.updatedInvoice.ReferenceInvoiceID == nil || *repo.updatedInvoice.ReferenceInvoiceID != 9999 {
			t.Errorf("expected ReferenceInvoiceID updated to 9999, got %v", repo.updatedInvoice.ReferenceInvoiceID)
		}
		if repo.updatedInvoice.Amount != 200.0 {
			t.Errorf("expected recalculated amount 200.0, got %.2f", repo.updatedInvoice.Amount)
		}
	})

	t.Run("fails when billing cancellation returns error", func(t *testing.T) {
		repo := newMockInvoiceUpdateRepoForService()
		logger := &mockLogger{}
		billingClient := &mockBillingClient{
			mockCancelErr: errors.New("cannot cancel paid invoice"),
		}

		billingID := int64(789)
		vendor := &domain.Vendor{ID: 1, Nickname: "vendedor_teste"}
		customer := &domain.Customer{ID: 10, VendorID: 1, Nickname: "cliente_teste", Vendor: vendor}
		plan := &domain.Plan{ID: 1, CustomerID: 10, Customer: customer}

		repo.invoices[1] = &domain.Invoice{
			ID:                 1,
			DueDate:            time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
			InvoicingDate:      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			Amount:             250.0,
			ReferenceInvoiceID: &billingID,
			PlanID:             1,
			Plan:               plan,
		}

		svc := NewInvoiceUpdate(repo, logger, billingClient)
		req := &dto.InvoiceUpdateRequest{
			InvoiceID: 1,
			DueDate:   "2026-11-05",
		}

		res := svc.Run(req)
		if res.GetStatusCode() != 500 {
			t.Fatalf("expected 500 when billing cancellation fails, got %d with message %s", res.GetStatusCode(), res.GetMessage())
		}
		if len(billingClient.createdInvoices) != 0 {
			t.Errorf("expected no invoice created when cancellation fails")
		}
	})
}
