package dto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInvoicingRequest_UnmarshalJSON(t *testing.T) {
	t.Run("English fields", func(t *testing.T) {
		payload := `{
			"vendor_id": 1,
			"start_date": "2026-08-01",
			"end_date": "2026-10-31",
			"plan_type": 1,
			"service_id": 10,
			"customer_id": 100,
			"reprocess": true,
			"n_months": 2
		}`
		var req InvoicingRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if req.VendorID == nil || *req.VendorID != 1 {
			t.Errorf("expected VendorID 1, got %v", req.VendorID)
		}
		if req.StartDate == nil || *req.StartDate != "2026-08-01" {
			t.Errorf("expected StartDate 2026-08-01, got %v", req.StartDate)
		}
		if req.EndDate == nil || *req.EndDate != "2026-10-31" {
			t.Errorf("expected EndDate 2026-10-31, got %v", req.EndDate)
		}
		if req.PlanType == nil || *req.PlanType != 1 {
			t.Errorf("expected PlanType 1, got %v", req.PlanType)
		}
		if req.ServiceID == nil || *req.ServiceID != 10 {
			t.Errorf("expected ServiceID 10, got %v", req.ServiceID)
		}
		if req.CustomerID == nil || *req.CustomerID != 100 {
			t.Errorf("expected CustomerID 100, got %v", req.CustomerID)
		}
		if req.Reprocess == nil || !*req.Reprocess {
			t.Errorf("expected Reprocess true, got %v", req.Reprocess)
		}
		if req.NMonths == nil || *req.NMonths != 2 {
			t.Errorf("expected NMonths 2, got %v", req.NMonths)
		}
	})

	t.Run("Portuguese fields", func(t *testing.T) {
		payload := `{
			"vendor": 2,
			"data_inicio": "2026-09-01",
			"data_fim": "2026-09-30",
			"tipo_plano": 2,
			"tipo_servico": 20,
			"cliente_id": 200,
			"reprocessar": true,
			"n_meses": 3
		}`
		var req InvoicingRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if req.VendorID == nil || *req.VendorID != 2 {
			t.Errorf("expected VendorID 2, got %v", req.VendorID)
		}
		if req.StartDate == nil || *req.StartDate != "2026-09-01" {
			t.Errorf("expected StartDate 2026-09-01, got %v", req.StartDate)
		}
		if req.EndDate == nil || *req.EndDate != "2026-09-30" {
			t.Errorf("expected EndDate 2026-09-30, got %v", req.EndDate)
		}
		if req.PlanType == nil || *req.PlanType != 2 {
			t.Errorf("expected PlanType 2, got %v", req.PlanType)
		}
		if req.ServiceID == nil || *req.ServiceID != 20 {
			t.Errorf("expected ServiceID 20, got %v", req.ServiceID)
		}
		if req.CustomerID == nil || *req.CustomerID != 200 {
			t.Errorf("expected CustomerID 200, got %v", req.CustomerID)
		}
		if req.Reprocess == nil || !*req.Reprocess {
			t.Errorf("expected Reprocess true, got %v", req.Reprocess)
		}
		if req.NMonths == nil || *req.NMonths != 3 {
			t.Errorf("expected NMonths 3, got %v", req.NMonths)
		}
	})
}

func TestInvoicingRequest_Validate(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		start := "2026-08-01"
		end := "2026-10-31"
		ptype := 1
		req := InvoicingRequest{
			StartDate: &start,
			EndDate:   &end,
			PlanType:  &ptype,
		}
		if err := req.Validate(nil); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}
	})

	t.Run("Invalid PlanType", func(t *testing.T) {
		ptype := 9
		req := InvoicingRequest{PlanType: &ptype}
		if err := req.Validate(nil); err == nil {
			t.Error("expected error for invalid plan_type")
		}
	})

	t.Run("Invalid StartDate format", func(t *testing.T) {
		badDate := "01/08/2026"
		req := InvoicingRequest{StartDate: &badDate}
		if err := req.Validate(nil); err == nil {
			t.Error("expected error for invalid start_date format")
		}
	})

	t.Run("Start date after end date", func(t *testing.T) {
		start := "2026-11-01"
		end := "2026-10-01"
		req := InvoicingRequest{StartDate: &start, EndDate: &end}
		if err := req.Validate(nil); err == nil {
			t.Error("expected error when start date is after end date")
		}
	})
}

func TestInvoicingRequest_GetParsedDates(t *testing.T) {
	now := time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC)

	t.Run("Default dates (1 month before current month start to current month end)", func(t *testing.T) {
		req := InvoicingRequest{}
		startDate, endDate, err := req.GetParsedDates(now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

		if !startDate.Equal(expectedStart) {
			t.Errorf("expected startDate %v, got %v", expectedStart, startDate)
		}
		if !endDate.Equal(expectedEnd) {
			t.Errorf("expected endDate %v, got %v", expectedEnd, endDate)
		}
	})

	t.Run("Custom NMonths (3 months before)", func(t *testing.T) {
		n := 3
		req := InvoicingRequest{NMonths: &n}
		startDate, endDate, err := req.GetParsedDates(now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

		if !startDate.Equal(expectedStart) {
			t.Errorf("expected startDate %v, got %v", expectedStart, startDate)
		}
		if !endDate.Equal(expectedEnd) {
			t.Errorf("expected endDate %v, got %v", expectedEnd, endDate)
		}
	})

	t.Run("Explicit dates", func(t *testing.T) {
		start := "2026-05-15"
		end := "2026-08-20"
		req := InvoicingRequest{StartDate: &start, EndDate: &end}
		startDate, endDate, err := req.GetParsedDates(now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedStart := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

		if !startDate.Equal(expectedStart) {
			t.Errorf("expected startDate %v, got %v", expectedStart, startDate)
		}
		if !endDate.Equal(expectedEnd) {
			t.Errorf("expected endDate %v, got %v", expectedEnd, endDate)
		}
	})
}
