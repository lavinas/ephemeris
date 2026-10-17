package dto

import (
	"encoding/json"
	"testing"
	"time"

	"planner/internal/domain"
)

type mockInvoiceUpdateRepo struct {
	mockPlanRepo
	invoice *domain.Invoice
}

func (m *mockInvoiceUpdateRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	if m.invoice != nil && m.invoice.ID == id {
		return m.invoice, nil
	}
	return nil, nil
}

func TestInvoiceUpdateRequest_UnmarshalJSON(t *testing.T) {
	raw := `{
		"id": 10,
		"due_date": "2026-11-15",
		"items": [
			{"id": 101, "invoiced_quantity": 4},
			{"id": 102, "invoiced_quantity": 2}
		]
	}`
	var req InvoiceUpdateRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.InvoiceID != 10 || req.DueDate != "2026-11-15" || len(req.Items) != 2 {
		t.Fatalf("unexpected unmarshal values: %+v", req)
	}
	if req.Items[0].ID != 101 || req.Items[0].InvoicedQuantity != 4 {
		t.Errorf("unexpected item[0]: %+v", req.Items[0])
	}
}

func TestInvoiceUpdateRequest_Validation(t *testing.T) {
	sampleInvoice := &domain.Invoice{
		ID:      10,
		DueDate: time.Now(),
		Items: []domain.InvoiceItem{
			{ID: 101, InvoicedQuantity: 2, Price: 50.0},
			{ID: 102, InvoicedQuantity: 1, Price: 100.0},
		},
	}
	repo := &mockInvoiceUpdateRepo{invoice: sampleInvoice}

	t.Run("valid request", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 10,
			DueDate:   "2026-11-20",
			Items: []InvoiceItemUpdateRequest{
				{ID: 101, InvoicedQuantity: 5},
			},
		}
		if err := req.Validate(repo); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}
	})

	t.Run("invalid invoice id", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 0,
			DueDate:   "2026-11-20",
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for invalid invoice id")
		}
	})

	t.Run("missing due date", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 10,
			DueDate:   "",
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for empty due date")
		}
	})

	t.Run("invalid due date format", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 10,
			DueDate:   "20-11-2026",
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for invalid due date format")
		}
	})

	t.Run("invoice not found", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 999,
			DueDate:   "2026-11-20",
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for non-existent invoice")
		}
	})

	t.Run("item not in invoice", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 10,
			DueDate:   "2026-11-20",
			Items: []InvoiceItemUpdateRequest{
				{ID: 999, InvoicedQuantity: 1},
			},
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for foreign item")
		}
	})

	t.Run("negative item quantity", func(t *testing.T) {
		req := &InvoiceUpdateRequest{
			InvoiceID: 10,
			DueDate:   "2026-11-20",
			Items: []InvoiceItemUpdateRequest{
				{ID: 101, InvoicedQuantity: -1},
			},
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for negative quantity")
		}
	})
}

func TestInvoiceUpdateRequest_Reset(t *testing.T) {
	req := &InvoiceUpdateRequest{
		InvoiceID: 10,
		DueDate:   "2026-11-20",
		Items:     []InvoiceItemUpdateRequest{{ID: 1, InvoicedQuantity: 2}},
	}
	req.Reset()
	if req.InvoiceID != 0 || req.DueDate != "" || req.Items != nil {
		t.Errorf("expected reset to clear fields, got: %+v", req)
	}
}
