package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"planner/internal/domain"
)

type mockInvoiceCreateRepo struct {
	mockPlanRepo
	invoiceExists bool
}

func (m *mockInvoiceCreateRepo) CheckInvoiceExists(customerID int64, invoiceDate, periodStart, periodEnd time.Time, excludeInvoiceID int64) (bool, error) {
	return m.invoiceExists, nil
}

func TestInvoiceCreateRequest_UnmarshalJSON(t *testing.T) {
	raw := `{
		"nickname": "ana_flavia",
		"invoice_date": "2026-10-22",
		"due_date": "2026-10-25",
		"period_start": "2026-10-01",
		"period_end": "2026-10-31",
		"reference_invoice_id": 520,
		"items": [
			{
				"service_id": 1,
				"price": 110.0,
				"description": "aula canto",
				"quantity": 2
			}
		]
	}`

	var req InvoiceCreateRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if req.Nickname != "ana_flavia" {
		t.Errorf("expected nickname ana_flavia, got %s", req.Nickname)
	}
	if req.ReferenceInvoiceID == nil || *req.ReferenceInvoiceID != 520 {
		t.Errorf("expected reference_invoice_id 520, got %v", req.ReferenceInvoiceID)
	}
	if len(req.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(req.Items))
	}
	if req.Items[0].ServiceID != 1 || req.Items[0].Price != 110.0 || req.Items[0].Quantity != 2 {
		t.Errorf("unexpected item values: %+v", req.Items[0])
	}
}

func TestInvoiceCreateRequest_Validate(t *testing.T) {
	baseRepo := newMockPlanRepo()
	baseRepo.customers["ana_flavia"] = &domain.Customer{ID: 101, Nickname: "ana_flavia", Name: "Ana Flávia", VendorID: 1}
	baseRepo.services[10] = &domain.Service{ID: 10, Name: "online/canto/60", VendorID: 1}

	repo := &mockInvoiceCreateRepo{
		mockPlanRepo: *baseRepo,
	}

	t.Run("valid request with defaults", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "ana_flavia",
			Items: []InvoiceItemCreateRequest{
				{
					ServiceID: 10,
					Price:     110.0,
					Quantity:  2,
				},
			},
		}

		if err := req.Validate(repo); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
		if req.InvoiceDate == "" || req.DueDate == "" || req.PeriodStart == "" || req.PeriodEnd == "" {
			t.Errorf("expected defaults to be populated: %+v", req)
		}
	})

	t.Run("missing nickname", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 10, Price: 110.0, Quantity: 2},
			},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "nickname") {
			t.Fatalf("expected nickname error, got: %v", err)
		}
	})

	t.Run("customer not found", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "cliente_inexistente",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 10, Price: 110.0, Quantity: 2},
			},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "não encontrado") {
			t.Fatalf("expected customer not found error, got: %v", err)
		}
	})

	t.Run("due date before invoice date", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname:    "ana_flavia",
			InvoiceDate: "2026-10-22",
			DueDate:     "2026-10-20",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 10, Price: 110.0, Quantity: 2},
			},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "vencimento") {
			t.Fatalf("expected due date error, got: %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "ana_flavia",
			Items:    []InvoiceItemCreateRequest{},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "1 ou mais itens") {
			t.Fatalf("expected 1 ou mais itens error, got: %v", err)
		}
	})

	t.Run("price zero or negative", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "ana_flavia",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 10, Price: 0, Quantity: 2},
			},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "maior que zero") {
			t.Fatalf("expected price error, got: %v", err)
		}
	})

	t.Run("quantity zero or negative", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "ana_flavia",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 10, Price: 110.0, Quantity: 0},
			},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "maior que zero") {
			t.Fatalf("expected quantity error, got: %v", err)
		}
	})

	t.Run("service not registered", func(t *testing.T) {
		req := &InvoiceCreateRequest{
			Nickname: "ana_flavia",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 9999, Price: 110.0, Quantity: 2},
			},
		}
		if err := req.Validate(repo); err == nil || !strings.Contains(err.Error(), "serviço registrado") {
			t.Fatalf("expected service error, got: %v", err)
		}
	})

	t.Run("duplicate invoice in database", func(t *testing.T) {
		dupRepo := &mockInvoiceCreateRepo{
			mockPlanRepo:  *baseRepo,
			invoiceExists: true,
		}
		req := &InvoiceCreateRequest{
			Nickname: "ana_flavia",
			Items: []InvoiceItemCreateRequest{
				{ServiceID: 10, Price: 110.0, Quantity: 2},
			},
		}
		if err := req.Validate(dupRepo); err == nil || !strings.Contains(err.Error(), "não se pode ter mais do que 1 fatura") {
			t.Fatalf("expected duplicate error, got: %v", err)
		}
	})
}
