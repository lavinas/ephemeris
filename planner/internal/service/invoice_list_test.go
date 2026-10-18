package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoiceListRepo struct {
	mockPlanRepoForService
	invoices     []domain.Invoice
	totalCount   int64
	totalAmount  float64
	lastPage     int
	lastPageSize int
	lastConds    map[string]interface{}
	lastOrderBy  []string
}

func (m *mockInvoiceListRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	if nickname != nil && *nickname == "carlos" {
		return []domain.Customer{{ID: 789, Nickname: "carlos"}}, nil
	}
	return nil, nil
}

func (m *mockInvoiceListRepo) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	m.lastPage = page
	m.lastPageSize = pageSize
	m.lastConds = conditions
	m.lastOrderBy = orderBy
	return m.invoices, m.totalCount, m.totalAmount, nil
}

func TestInvoiceList_Success(t *testing.T) {
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	itemDesc := "Sessão de Fisioterapia"

	invoices := []domain.Invoice{
		{
			ID:            101,
			PlanID:        5,
			InvoicingDate: now,
			DueDate:       dueDate,
			Amount:        250.00,
			Plan: &domain.Plan{
				ID:       5,
				PlanType: 2,
				Customer: &domain.Customer{
					ID:       1,
					Nickname: "carlos",
				},
			},
			Items: []domain.InvoiceItem{
				{
					ID:               1,
					ServiceID:        2,
					Price:            125.00,
					Description:      &itemDesc,
					InvoicedQuantity: 2,
					RealizedQuantity: 1,
				},
			},
		},
	}

	repo := &mockInvoiceListRepo{
		invoices:    invoices,
		totalCount:  1,
		totalAmount: 250.00,
	}
	logger := &mockLogger{}

	svc := NewInvoiceList(repo, logger)
	req := &dto.InvoiceListRequest{
		Page:     1,
		PageSize: 10,
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d", res.GetStatusCode())
	}

	listResp, ok := res.(*dto.InvoiceListResponse)
	if !ok {
		t.Fatalf("expected *dto.InvoiceListResponse, got %T", res)
	}

	if listResp.TotalCount != 1 || listResp.TotalPages != 1 || len(listResp.Invoices) != 1 || listResp.TotalAmount != 250.00 {
		t.Fatalf("unexpected counts: total=%d, pages=%d, len=%d, totalAmount=%f", listResp.TotalCount, listResp.TotalPages, len(listResp.Invoices), listResp.TotalAmount)
	}

	inv := listResp.Invoices[0]
	if inv.ID != 101 || inv.InvoiceDate != "2026-09-15" || inv.DueDate != "2026-09-25" || inv.Nickname != "carlos" || inv.Amount != 250.00 {
		t.Errorf("invoice fields mismatch: %+v", inv)
	}
	if inv.PlanType != 2 || inv.PlanTypeName != "Pacote" {
		t.Errorf("expected plan_type=2 and plan_type_name='Pacote', got %d and %q", inv.PlanType, inv.PlanTypeName)
	}

	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}

	item := inv.Items[0]
	if item.ID != 1 || item.ServiceID != 2 || item.Description != "Sessão de Fisioterapia" || item.Price != 125.00 || item.InvoicedQuantity != 2 || item.RealizedQuantity != 1 {
		t.Errorf("item fields mismatch: %+v", item)
	}
}

func TestInvoiceList_ValidationFailure(t *testing.T) {
	repo := &mockInvoiceListRepo{}
	logger := &mockLogger{}

	svc := NewInvoiceList(repo, logger)
	req := &dto.InvoiceListRequest{
		InvoiceID: -5,
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 {
		t.Fatalf("expected 400, got %d", res.GetStatusCode())
	}
}

func TestInvoiceList_FiltersApplied(t *testing.T) {
	repo := &mockInvoiceListRepo{}
	logger := &mockLogger{}

	svc := NewInvoiceList(repo, logger)
	req := &dto.InvoiceListRequest{
		Page:      2,
		PageSize:  5,
		InvoiceID: 123,
		PlanID:    456,
		StartDate: "2026-08-01",
		EndDate:   "2026-08-31",
		Nickname:  "carlos",
		PlanType:  1,
		ServiceID: 10,
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d", res.GetStatusCode())
	}

	if repo.lastPage != 2 || repo.lastPageSize != 5 {
		t.Errorf("pagination not passed: page=%d, pageSize=%d", repo.lastPage, repo.lastPageSize)
	}

	if repo.lastConds["planner.invoice.id = ?"] != int64(123) {
		t.Errorf("expected invoice_id condition")
	}
	if repo.lastConds["planner.invoice.plan_id = ?"] != int64(456) {
		t.Errorf("expected plan_id condition")
	}
	if repo.lastConds["planner.invoice.plan_id IN (SELECT id FROM planner.plan WHERE customer_id = ? AND deleted_at IS NULL)"] != int64(789) {
		t.Errorf("expected customer condition")
	}
	if repo.lastConds["planner.invoice.plan_id IN (SELECT id FROM planner.plan WHERE plan_type = ? AND deleted_at IS NULL)"] != 1 {
		t.Errorf("expected plan_type condition")
	}
	if repo.lastConds["planner.invoice.id IN (SELECT invoice_id FROM planner.invoice_item WHERE service_id = ? AND deleted_at IS NULL)"] != int64(10) {
		t.Errorf("expected service_id condition")
	}

	// Test PendingOnly true
	req.PendingOnly = true
	res = svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d", res.GetStatusCode())
	}
	if _, ok := repo.lastConds["planner.invoice.reference_invoice_id IS NULL"]; !ok {
		t.Errorf("expected reference_invoice_id IS NULL condition when PendingOnly is true")
	}

	// Test PendingOnly false
	req.PendingOnly = false
	res = svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d", res.GetStatusCode())
	}
	if _, ok := repo.lastConds["planner.invoice.reference_invoice_id IS NULL"]; ok {
		t.Errorf("expected no reference_invoice_id condition when PendingOnly is false")
	}
}
