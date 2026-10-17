package dto

import (
	"encoding/json"
	"testing"
)

func TestInvoiceBatchDeleteRequest_UnmarshalJSON(t *testing.T) {
	raw := `{
		"start_date": "2026-10-01",
		"end_date": "2026-10-31",
		"invoice_id": 55,
		"plano_id": 12,
		"cliente": "pedro",
		"tipo_filtro": 1,
		"servico_filtro": 7
	}`
	var req InvoiceBatchDeleteRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.StartDate != "2026-10-01" || req.EndDate != "2026-10-31" || req.InvoiceID != 55 ||
		req.PlanID != 12 || req.Nickname != "pedro" || req.PlanType != 1 || req.ServiceID != 7 {
		t.Fatalf("unexpected unmarshal values: %+v", req)
	}
}

func TestInvoiceBatchDeleteRequest_Validation(t *testing.T) {
	repo := &mockInvoiceRepoForDTO{}

	t.Run("valid request with all fields", func(t *testing.T) {
		req := &InvoiceBatchDeleteRequest{
			StartDate: "2026-10-01",
			EndDate:   "2026-10-31",
			Nickname:  "joao",
			PlanType:  1,
			ServiceID: 5,
		}
		if err := req.Validate(repo); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}
		if len(req.CustomerIDs) != 1 || req.CustomerIDs[0] != 42 {
			t.Errorf("expected CustomerIDs [42], got: %v", req.CustomerIDs)
		}
	})

	t.Run("invalid date range", func(t *testing.T) {
		req := &InvoiceBatchDeleteRequest{
			StartDate: "2026-10-31",
			EndDate:   "2026-10-01",
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error when end date is before start date")
		}
	})

	t.Run("invalid plan type", func(t *testing.T) {
		req := &InvoiceBatchDeleteRequest{
			PlanType: 9,
		}
		if err := req.Validate(repo); err == nil {
			t.Error("expected error for invalid plan type")
		}
	})

	t.Run("customer not found", func(t *testing.T) {
		req := &InvoiceBatchDeleteRequest{
			Nickname: "inexistente",
		}
		if err := req.Validate(repo); err != nil {
			t.Errorf("expected valid even if nickname not found, got error: %v", err)
		}
		if len(req.CustomerIDs) != 1 || req.CustomerIDs[0] != -1 {
			t.Errorf("expected CustomerIDs [-1], got: %v", req.CustomerIDs)
		}
	})
}

func TestInvoiceBatchDeleteRequest_Reset(t *testing.T) {
	req := &InvoiceBatchDeleteRequest{
		StartDate: "2026-10-01",
		Nickname:  "test",
		PlanType:  2,
	}
	req.Reset()
	if req.StartDate != "" || req.Nickname != "" || req.PlanType != 0 {
		t.Errorf("expected reset to clear fields, got: %+v", req)
	}
}
