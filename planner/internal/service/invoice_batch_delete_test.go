package service

import (
	"errors"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoiceBatchDeleteRepo struct {
	mockPlanRepoForService
	invoices   []domain.Invoice
	deletedIDs []int64
	failFind   bool
	failDelete bool
}

func (m *mockInvoiceBatchDeleteRepo) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	if m.failFind {
		return nil, errors.New("database find error")
	}
	var res []domain.Invoice
	for _, inv := range m.invoices {
		if inv.DeletedAt == nil {
			res = append(res, inv)
		}
	}
	return res, nil
}

func (m *mockInvoiceBatchDeleteRepo) DeleteInvoicesByIDs(ids []int64) error {
	if m.failDelete {
		return errors.New("database delete error")
	}
	m.deletedIDs = append(m.deletedIDs, ids...)
	now := time.Now()
	for _, id := range ids {
		for i := range m.invoices {
			if m.invoices[i].ID == id {
				m.invoices[i].DeletedAt = &now
			}
		}
	}
	return nil
}

func TestInvoiceBatchDelete_Success(t *testing.T) {
	repo := &mockInvoiceBatchDeleteRepo{
		invoices: []domain.Invoice{
			{ID: 101, Amount: 150.0},
			{ID: 102, Amount: 200.0},
			{ID: 103, Amount: 350.0},
		},
	}
	logger := &mockLogger{}

	svc := NewInvoiceBatchDelete(repo, logger)
	req := &dto.InvoiceBatchDeleteRequest{}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d: %s", res.GetStatusCode(), res.GetMessage())
	}

	batchResp, ok := res.(*dto.InvoiceBatchDeleteResponse)
	if !ok {
		t.Fatalf("expected InvoiceBatchDeleteResponse")
	}
	if batchResp.DeletedCount != 3 {
		t.Errorf("expected 3 deleted invoices, got %d", batchResp.DeletedCount)
	}
	if len(repo.deletedIDs) != 3 {
		t.Errorf("expected repo to receive 3 deleted IDs, got %d", len(repo.deletedIDs))
	}
}

func TestInvoiceBatchDelete_NoInvoicesFound(t *testing.T) {
	repo := &mockInvoiceBatchDeleteRepo{
		invoices: []domain.Invoice{},
	}
	logger := &mockLogger{}

	svc := NewInvoiceBatchDelete(repo, logger)
	req := &dto.InvoiceBatchDeleteRequest{
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d: %s", res.GetStatusCode(), res.GetMessage())
	}

	batchResp := res.(*dto.InvoiceBatchDeleteResponse)
	if batchResp.DeletedCount != 0 {
		t.Errorf("expected 0 deleted invoices, got %d", batchResp.DeletedCount)
	}
}

func TestInvoiceBatchDelete_ValidationFailure(t *testing.T) {
	repo := &mockInvoiceBatchDeleteRepo{}
	logger := &mockLogger{}

	svc := NewInvoiceBatchDelete(repo, logger)
	req := &dto.InvoiceBatchDeleteRequest{
		StartDate: "2026-10-31",
		EndDate:   "2026-10-01", // end date before start date
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 {
		t.Errorf("expected 400 on validation failure, got %d", res.GetStatusCode())
	}
}

func TestInvoiceBatchDelete_RepositoryError(t *testing.T) {
	repo := &mockInvoiceBatchDeleteRepo{
		invoices: []domain.Invoice{
			{ID: 101, Amount: 150.0},
		},
		failDelete: true,
	}
	logger := &mockLogger{}

	svc := NewInvoiceBatchDelete(repo, logger)
	req := &dto.InvoiceBatchDeleteRequest{}

	res := svc.Run(req)
	if res.GetStatusCode() != 500 {
		t.Errorf("expected 500 on repository error, got %d", res.GetStatusCode())
	}
}
