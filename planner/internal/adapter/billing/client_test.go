package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"planner/internal/port"
)

func TestClient_CreateInvoice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/invoice/api/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req billingCreateRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(req.Items))
		}
		if req.Items[0].Customer != "cliente1" || req.Items[0].Vendor != "vendedor1" {
			t.Errorf("unexpected payload: %+v", req.Items[0])
		}

		resp := billingResponseBase{
			HttpCode: 200,
			Status:   "success",
			Message:  "Invoice created successfully",
			ID:       101,
			IDs:      []int64{101},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, server.Client())
	extCode := "50"
	id, err := client.CreateInvoice(context.Background(), port.BillingInvoiceInput{
		Vendor:       "vendedor1",
		Customer:     "cliente1",
		Invoicing:    "2026-10-17",
		Due:          "2026-10-25",
		ExternalCode: &extCode,
		Items: []port.BillingItemInput{
			{Description: "Item 1", Price: 150.0, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 101 {
		t.Errorf("expected invoice ID 101, got %d", id)
	}
}

func TestClient_CreateInvoice_DuplicateRecovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/invoice/api/create":
			// Return duplicate external code error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(billingResponseBase{
				HttpCode: 400,
				Status:   "error",
				Message:  "invoice 0: duplicate external_code '50' for vendor",
			})
		case "/invoice/api/list":
			vendor := r.URL.Query().Get("vendor")
			extCode := r.URL.Query().Get("external_code")
			if vendor != "vendedor1" || extCode != "50" {
				t.Errorf("unexpected list query: vendor=%s, ext=%s", vendor, extCode)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"http_code": 200,
				"status":    "success",
				"invoices": []map[string]interface{}{
					{
						"id":            202,
						"customer":      "cliente1",
						"external_code": "50",
					},
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, server.Client())
	extCode := "50"
	id, err := client.CreateInvoice(context.Background(), port.BillingInvoiceInput{
		Vendor:       "vendedor1",
		Customer:     "cliente1",
		Invoicing:    "2026-10-17",
		Due:          "2026-10-25",
		ExternalCode: &extCode,
		Items: []port.BillingItemInput{
			{Description: "Item 1", Price: 150.0, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected duplicate to be recovered, got error: %v", err)
	}
	if id != 202 {
		t.Errorf("expected recovered invoice ID 202, got %d", id)
	}
}

func TestClient_GetInvoiceByExternalCode_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"http_code": 200,
			"status":    "success",
			"invoices":  []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, server.Client())
	id, err := client.GetInvoiceByExternalCode(context.Background(), "vend", "none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != nil {
		t.Errorf("expected nil for not found, got %v", id)
	}
}
