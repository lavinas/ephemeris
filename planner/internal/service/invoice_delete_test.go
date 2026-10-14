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
