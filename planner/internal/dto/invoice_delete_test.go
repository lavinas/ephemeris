package dto

import (
	"encoding/json"
	"testing"
)

func TestInvoiceDeleteRequest_UnmarshalJSON(t *testing.T) {
	t.Run("Standard id field", func(t *testing.T) {
		payload := `{"id": 123}`
		var req InvoiceDeleteRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if req.InvoiceID != 123 {
			t.Errorf("expected InvoiceID 123, got %d", req.InvoiceID)
		}
	})

	t.Run("invoice_id field", func(t *testing.T) {
		payload := `{"invoice_id": 456}`
		var req InvoiceDeleteRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if req.InvoiceID != 456 {
			t.Errorf("expected InvoiceID 456, got %d", req.InvoiceID)
		}
	})

	t.Run("fatura_id field", func(t *testing.T) {
		payload := `{"fatura_id": 789}`
		var req InvoiceDeleteRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if req.InvoiceID != 789 {
			t.Errorf("expected InvoiceID 789, got %d", req.InvoiceID)
		}
	})
}

func TestInvoiceDeleteRequest_ValidateAndReset(t *testing.T) {
	t.Run("Valid id", func(t *testing.T) {
		req := InvoiceDeleteRequest{InvoiceID: 10}
		if err := req.Validate(nil); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}
		req.Reset()
		if req.InvoiceID != 0 {
			t.Errorf("expected reset to 0, got %d", req.InvoiceID)
		}
	})

	t.Run("Invalid zero id", func(t *testing.T) {
		req := InvoiceDeleteRequest{InvoiceID: 0}
		if err := req.Validate(nil); err == nil {
			t.Error("expected error for zero id")
		}
	})

	t.Run("Invalid negative id", func(t *testing.T) {
		req := InvoiceDeleteRequest{InvoiceID: -5}
		if err := req.Validate(nil); err == nil {
			t.Error("expected error for negative id")
		}
	})
}

func TestInvoiceDeleteResponse(t *testing.T) {
	resp := NewInvoiceDeleteResponse(200, "OK", "fatura excluída com sucesso")
	if resp.GetStatusCode() != 200 {
		t.Errorf("expected status 200, got %d", resp.GetStatusCode())
	}
	if resp.GetMessage() != "fatura excluída com sucesso" {
		t.Errorf("expected message 'fatura excluída com sucesso', got '%s'", resp.GetMessage())
	}
}
