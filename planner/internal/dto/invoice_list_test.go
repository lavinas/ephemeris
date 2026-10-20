package dto

import (
	"encoding/json"
	"testing"

	"planner/internal/domain"
)

type mockInvoiceRepoForDTO struct {
	mockPlanRepo
	customers []domain.Customer
}

func (m *mockInvoiceRepoForDTO) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	if nickname != nil && *nickname == "inexistente" {
		return []domain.Customer{}, nil
	}
	return []domain.Customer{
		{ID: 42, Nickname: "joao"},
	}, nil
}

func TestInvoiceListRequest_Defaults(t *testing.T) {
	req := &InvoiceListRequest{}
	if err := req.Validate(nil); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if req.Page != 1 {
		t.Errorf("expected page 1, got %d", req.Page)
	}
	if req.PageSize != 10 {
		t.Errorf("expected page_size 10, got %d", req.PageSize)
	}
}

func TestInvoiceListRequest_UnmarshalJSON(t *testing.T) {
	raw := `{
		"page": 2,
		"page_size": 25,
		"start_date": "2026-09-01",
		"end_date": "2026-09-30",
		"invoice_id": 105,
		"nickname": "maria",
		"plan_type": 1,
		"service_id": 8,
		"plan_id": 15
	}`
	var req InvoiceListRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.Page != 2 || req.PageSize != 25 || req.InvoiceID != 105 || req.PlanID != 15 || req.ServiceID != 8 || req.PlanType != 1 || req.Nickname != "maria" || req.StartDate != "2026-09-01" || req.EndDate != "2026-09-30" {
		t.Errorf("unmarshaled request mismatch: %+v", req)
	}
}

func TestInvoiceListRequest_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     InvoiceListRequest
		wantErr bool
	}{
		{
			name:    "negative invoice id",
			req:     InvoiceListRequest{InvoiceID: -1},
			wantErr: true,
		},
		{
			name:    "negative plan id",
			req:     InvoiceListRequest{PlanID: -1},
			wantErr: true,
		},
		{
			name:    "invalid plan type",
			req:     InvoiceListRequest{PlanType: 5},
			wantErr: true,
		},
		{
			name:    "negative service id",
			req:     InvoiceListRequest{ServiceID: -1},
			wantErr: true,
		},
		{
			name:    "invalid start date format",
			req:     InvoiceListRequest{StartDate: "01-09-2026"},
			wantErr: true,
		},
		{
			name:    "start date after end date",
			req:     InvoiceListRequest{StartDate: "2026-10-01", EndDate: "2026-09-01"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInvoiceListRequest_NicknameResolution(t *testing.T) {
	repo := &mockInvoiceRepoForDTO{}

	// Existing nickname
	req := InvoiceListRequest{Nickname: "joao"}
	if err := req.Validate(repo); err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if len(req.CustomerIDs) != 1 || req.CustomerIDs[0] != 42 {
		t.Errorf("expected customerIDs [42], got %v", req.CustomerIDs)
	}

	// Non-existing nickname
	req2 := InvoiceListRequest{Nickname: "inexistente"}
	if err := req2.Validate(repo); err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if len(req2.CustomerIDs) != 1 || req2.CustomerIDs[0] != -1 {
		t.Errorf("expected customerIDs [-1], got %v", req2.CustomerIDs)
	}
}

func TestInvoiceListResponse_TotalAmount(t *testing.T) {
	invoices := []InvoiceDTO{
		{
			ID:          1,
			InvoiceDate: "2026-09-01",
			Nickname:    "joao",
			DueDate:     "2026-09-10",
			Amount:      150.50,
		},
		{
			ID:          2,
			InvoiceDate: "2026-09-02",
			Nickname:    "maria",
			DueDate:     "2026-09-12",
			Amount:      250.00,
		},
	}
	resp := NewInvoiceListResponse(200, "success", "ok", invoices, 1, 10, 1, 2, 400.50)
	if resp.TotalAmount != 400.50 {
		t.Errorf("expected TotalAmount 400.50, got %f", resp.TotalAmount)
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", resp.TotalCount)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if raw["total_amount"] != 400.50 {
		t.Errorf("expected total_amount 400.50 in JSON, got %v", raw["total_amount"])
	}
}

func TestInvoiceListRequest_PendingOnly(t *testing.T) {
	testCases := []struct {
		json     string
		expected bool
	}{
		{`{"pending_only": true}`, true},
		{`{"pendente": "true"}`, true},
		{`{"pending": "1"}`, true},
		{`{"pending_only": false}`, false},
		{`{"pendente": "false"}`, false},
		{`{"pendente": "0"}`, false},
		{`{}`, false},
	}

	for _, tc := range testCases {
		var req InvoiceListRequest
		if err := json.Unmarshal([]byte(tc.json), &req); err != nil {
			t.Fatalf("failed to unmarshal %s: %v", tc.json, err)
		}
		if req.PendingOnly != tc.expected {
			t.Errorf("for %s expected PendingOnly=%v, got %v", tc.json, tc.expected, req.PendingOnly)
		}
	}
}
