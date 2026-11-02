package dto

import (
	"encoding/json"
	"testing"
)

func TestInvoiceItemAddRequest_UnmarshalAndValidate(t *testing.T) {
	t.Run("Addition with English fields", func(t *testing.T) {
		raw := `{
			"invoice_id": 10,
			"service_id": 2,
			"operation": "add",
			"period_start": "2026-10-01",
			"period_end": "2026-10-31",
			"price": 120.50,
			"quantity": 2,
			"description": "Sessões extras"
		}`
		var req InvoiceItemAddRequest
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if err := req.Validate(nil); err != nil {
			t.Fatalf("unexpected validate error: %v", err)
		}
		if req.Price != 120.50 {
			t.Errorf("expected positive price 120.50, got %v", req.Price)
		}
		if req.Operation != "add" {
			t.Errorf("expected operation 'add', got %q", req.Operation)
		}
	})

	t.Run("Discount with Portuguese fields converts price to negative", func(t *testing.T) {
		raw := `{
			"fatura_id": "15",
			"servico": "Psicoterapia",
			"operacao": "desconto",
			"inicio": "01/10/2026",
			"fim": "31/10/2026",
			"preco": "100,00",
			"quantidade": "1"
		}`
		var req InvoiceItemAddRequest
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if err := req.Validate(nil); err != nil {
			t.Fatalf("unexpected validate error: %v", err)
		}
		if req.Price != -100.00 {
			t.Errorf("expected negative price -100.00 for discount, got %v", req.Price)
		}
		if req.Operation != "discount" {
			t.Errorf("expected operation 'discount', got %q", req.Operation)
		}
	})

	t.Run("Validation errors for invalid period, zero price or zero quantity", func(t *testing.T) {
		req := InvoiceItemAddRequest{
			InvoiceID:   1,
			ServiceID:   1,
			PeriodStart: "2026-10-31",
			PeriodEnd:   "2026-10-01",
			Price:       100,
			Quantity:    1,
		}
		if err := req.Validate(nil); err == nil {
			t.Errorf("expected error when period_end < period_start")
		}

		req2 := InvoiceItemAddRequest{
			InvoiceID:   1,
			ServiceID:   1,
			PeriodStart: "2026-10-01",
			PeriodEnd:   "2026-10-31",
			Price:       0,
			Quantity:    1,
		}
		if err := req2.Validate(nil); err == nil {
			t.Errorf("expected error when price is 0")
		}
	})
}

func TestInvoiceItemDeleteRequest_UnmarshalAndValidate(t *testing.T) {
	raw := `{"item_id": 42, "invoice_id": 7}`
	var req InvoiceItemDeleteRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if req.ItemID != 42 || req.InvoiceID != 7 {
		t.Errorf("expected ItemID=42 InvoiceID=7, got %d / %d", req.ItemID, req.InvoiceID)
	}
	if err := req.Validate(nil); err != nil {
		t.Errorf("unexpected validate error: %v", err)
	}
}

