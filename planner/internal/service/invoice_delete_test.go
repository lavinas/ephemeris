package service

import (
	"errors"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoiceDeleteRepo struct {
	mockPlanRepoForService
	invoices   map[int64]*domain.Invoice
	deletedIDs []int64
	failGet    bool
	failDelete bool
}

func newMockInvoiceDeleteRepo() *mockInvoiceDeleteRepo {
	return &mockInvoiceDeleteRepo{
		invoices:   make(map[int64]*domain.Invoice),
		deletedIDs: make([]int64, 0),
	}
}

func (m *mockInvoiceDeleteRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	if m.failGet {
		return nil, errors.New("database connection error")
	}
	inv, ok := m.invoices[id]
	if !ok || inv.DeletedAt != nil {
		return nil, nil
	}
	return inv, nil
}

func (m *mockInvoiceDeleteRepo) DeleteInvoice(id int64) error {
	if m.failDelete {
		return errors.New("failed to delete invoice")
	}
	if inv, ok := m.invoices[id]; ok {
		now := time.Now()
		inv.DeletedAt = &now
		m.deletedIDs = append(m.deletedIDs, id)
	}
	return nil
}

func (m *mockInvoiceDeleteRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "vendedor_teste"}, nil
}

func (m *mockInvoiceDeleteRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, VendorID: 1, Nickname: "cliente_teste"}, nil
}

func TestInvoiceDelete_Success(t *testing.T) {
	repo := newMockInvoiceDeleteRepo()
	logger := &mockLogger{}
	repo.invoices[100] = &domain.Invoice{
		ID:     100,
		PlanID: 1,
		Amount: 250.0,
	}

	svc := NewInvoiceDelete(repo, logger)
	req := &dto.InvoiceDeleteRequest{InvoiceID: 100}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d with message %s", res.GetStatusCode(), res.GetMessage())
	}
	if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 100 {
		t.Errorf("expected invoice 100 in deletedIDs, got %v", repo.deletedIDs)
	}
}

func TestInvoiceDelete_NotFound(t *testing.T) {
	repo := newMockInvoiceDeleteRepo()
	logger := &mockLogger{}

	svc := NewInvoiceDelete(repo, logger)
	req := &dto.InvoiceDeleteRequest{InvoiceID: 999}

	res := svc.Run(req)
	if res.GetStatusCode() != 404 {
		t.Fatalf("expected status 404, got %d", res.GetStatusCode())
	}
}

func TestInvoiceDelete_ValidationError(t *testing.T) {
	repo := newMockInvoiceDeleteRepo()
	logger := &mockLogger{}

	svc := NewInvoiceDelete(repo, logger)
	req := &dto.InvoiceDeleteRequest{InvoiceID: 0}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 {
		t.Fatalf("expected status 400, got %d", res.GetStatusCode())
	}
}

func TestInvoiceDelete_RepoErrors(t *testing.T) {
	t.Run("Get error", func(t *testing.T) {
		repo := newMockInvoiceDeleteRepo()
		repo.failGet = true
		logger := &mockLogger{}

		svc := NewInvoiceDelete(repo, logger)
		req := &dto.InvoiceDeleteRequest{InvoiceID: 100}

		res := svc.Run(req)
		if res.GetStatusCode() != 500 {
			t.Fatalf("expected status 500, got %d", res.GetStatusCode())
		}
	})

	t.Run("Delete error", func(t *testing.T) {
		repo := newMockInvoiceDeleteRepo()
		repo.invoices[100] = &domain.Invoice{ID: 100}
		repo.failDelete = true
		logger := &mockLogger{}

		svc := NewInvoiceDelete(repo, logger)
		req := &dto.InvoiceDeleteRequest{InvoiceID: 100}

		res := svc.Run(req)
		if res.GetStatusCode() != 500 {
			t.Fatalf("expected status 500, got %d", res.GetStatusCode())
		}
	})
}

func TestInvoiceDelete_AlreadyInvoiced(t *testing.T) {
	t.Run("successfully cancels in billing and deletes in planner", func(t *testing.T) {
		repo := newMockInvoiceDeleteRepo()
		logger := &mockLogger{}
		billingClient := &mockBillingClient{}
		billingID := int64(456)

		vendor := &domain.Vendor{ID: 1, Nickname: "vendedor_teste"}
		customer := &domain.Customer{ID: 10, VendorID: 1, Nickname: "cliente_teste", Vendor: vendor}
		plan := &domain.Plan{ID: 1, CustomerID: 10, Customer: customer}

		repo.invoices[100] = &domain.Invoice{
			ID:                 100,
			PlanID:             1,
			Plan:               plan,
			Amount:             250.0,
			ReferenceInvoiceID: &billingID,
		}

		svc := NewInvoiceDelete(repo, logger, billingClient)
		req := &dto.InvoiceDeleteRequest{InvoiceID: 100}

		res := svc.Run(req)
		if res.GetStatusCode() != 200 {
			t.Fatalf("expected status 200, got %d with message %s", res.GetStatusCode(), res.GetMessage())
		}
		if len(billingClient.cancelledInvoices) != 1 {
			t.Fatalf("expected 1 cancelled invoice in billing, got %d", len(billingClient.cancelledInvoices))
		}
		cancelled := billingClient.cancelledInvoices[0]
		if cancelled.ID != 456 {
			t.Errorf("expected cancelled billing ID 456, got %d", cancelled.ID)
		}
		if cancelled.Vendor != "vendedor_teste" {
			t.Errorf("expected vendor 'vendedor_teste', got '%s'", cancelled.Vendor)
		}
		todayStr := time.Now().Format("2006-01-02")
		if cancelled.CancellationDate == nil || *cancelled.CancellationDate != todayStr {
			t.Errorf("expected cancellation date %s, got %v", todayStr, cancelled.CancellationDate)
		}
		if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != 100 {
			t.Errorf("expected invoice 100 to be deleted in repo, got %v", repo.deletedIDs)
		}
	})

	t.Run("fails when billing cancellation returns error", func(t *testing.T) {
		repo := newMockInvoiceDeleteRepo()
		logger := &mockLogger{}
		billingClient := &mockBillingClient{
			mockCancelErr: errors.New("billing service unavailable"),
		}
		billingID := int64(456)

		vendor := &domain.Vendor{ID: 1, Nickname: "vendedor_teste"}
		customer := &domain.Customer{ID: 10, VendorID: 1, Nickname: "cliente_teste", Vendor: vendor}
		plan := &domain.Plan{ID: 1, CustomerID: 10, Customer: customer}

		repo.invoices[100] = &domain.Invoice{
			ID:                 100,
			PlanID:             1,
			Plan:               plan,
			Amount:             250.0,
			ReferenceInvoiceID: &billingID,
		}

		svc := NewInvoiceDelete(repo, logger, billingClient)
		req := &dto.InvoiceDeleteRequest{InvoiceID: 100}

		res := svc.Run(req)
		if res.GetStatusCode() != 500 {
			t.Fatalf("expected status 500, got %d with message %s", res.GetStatusCode(), res.GetMessage())
		}
		if len(repo.deletedIDs) != 0 {
			t.Errorf("expected no invoice to be deleted in repo upon billing cancel failure, got %v", repo.deletedIDs)
		}
	})
}
