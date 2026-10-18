package dto

import (
	"strings"
	"testing"
	"time"

	"billing/internal/domain"
)

func TestInvoiceCancel_Validation(t *testing.T) {
	vendor := &domain.Vendor{ID: 10, Nickname: "estudio_amelia"}
	customer := &domain.Customer{ID: 20, VendorID: 10, Nickname: "cliente_1"}
	now := time.Now()
	cancelledDate := now.Add(-24 * time.Hour)
	paidDate := now.Add(-12 * time.Hour)
	taxDate := now.Add(-6 * time.Hour)

	extCode := "EXT-CANCEL-1"
	extCode2 := "EXT-CANCEL-2"
	repo := &invoiceCreateMockRepo{
		vendor:   vendor,
		customer: customer,
		invoices: map[string]*domain.Invoice{
			extCode: {
				ID:           101,
				VendorID:     10,
				Customer:     domain.Customer{ID: 20, VendorID: 10},
				ExternalCode: &extCode,
			},
			extCode2: {
				ID:           105,
				VendorID:     10,
				Customer:     domain.Customer{ID: 20, VendorID: 10},
				ExternalCode: &extCode2,
			},
			"EXT-ALREADY-CANCELLED": {
				ID:               102,
				VendorID:         10,
				Customer:         domain.Customer{ID: 20, VendorID: 10},
				CancellationDate: &cancelledDate,
			},
			"EXT-ALREADY-PAID": {
				ID:          103,
				VendorID:    10,
				Customer:    domain.Customer{ID: 20, VendorID: 10},
				PaymentDate: &paidDate,
			},
			"EXT-TAX-ISSUED": {
				ID:       104,
				VendorID: 10,
				Customer: domain.Customer{ID: 20, VendorID: 10},
				TaxDate:  &taxDate,
			},
		},
	}

	t.Run("valid cancellation by ID without date -> defaults to today", func(t *testing.T) {
		req := &InvoiceCancelRequest{
			Vendor: "estudio_amelia",
			ID:     101,
		}
		if err := req.Validate(repo); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		inv := req.GetDomain()
		if inv == nil || inv.CancellationDate == nil {
			t.Fatalf("expected CancellationDate to be set")
		}
	})

	t.Run("valid cancellation by ExternalCode with explicit date", func(t *testing.T) {
		code := "EXT-CANCEL-2"
		dateStr := "2026-10-18"
		req := &InvoiceCancelRequest{
			Vendor:           "estudio_amelia",
			ExternalCode:     &code,
			CancellationDate: &dateStr,
		}
		if err := req.Validate(repo); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
		inv := req.GetDomain()
		if inv == nil || inv.CancellationDate == nil {
			t.Fatalf("expected CancellationDate to be set")
		}
		if inv.CancellationDate.Format("2006-01-02") != "2026-10-18" {
			t.Errorf("expected cancellation date 2026-10-18, got %s", inv.CancellationDate.Format("2006-01-02"))
		}
	})

	t.Run("missing vendor", func(t *testing.T) {
		req := &InvoiceCancelRequest{
			ID: 101,
		}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "vendor is required") {
			t.Errorf("expected 'vendor is required' error, got %v", err)
		}
	})

	t.Run("missing both id and external_code", func(t *testing.T) {
		req := &InvoiceCancelRequest{
			Vendor: "estudio_amelia",
		}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "id or external_code is required") {
			t.Errorf("expected 'id or external_code is required' error, got %v", err)
		}
	})

	t.Run("already cancelled invoice", func(t *testing.T) {
		req := &InvoiceCancelRequest{
			Vendor: "estudio_amelia",
			ID:     102,
		}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "already cancelled") {
			t.Errorf("expected 'already cancelled' error, got %v", err)
		}
	})

	t.Run("already paid invoice", func(t *testing.T) {
		req := &InvoiceCancelRequest{
			Vendor: "estudio_amelia",
			ID:     103,
		}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "already paid") {
			t.Errorf("expected 'already paid' error, got %v", err)
		}
	})

	t.Run("tax already issued", func(t *testing.T) {
		req := &InvoiceCancelRequest{
			Vendor: "estudio_amelia",
			ID:     104,
		}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "tax document already issued") {
			t.Errorf("expected 'tax document already issued' error, got %v", err)
		}
	})

	t.Run("invalid cancellation date format", func(t *testing.T) {
		invalidDate := "not-a-date"
		req := &InvoiceCancelRequest{
			Vendor:           "estudio_amelia",
			ID:               101,
			CancellationDate: &invalidDate,
		}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "invalid cancellation date") {
			t.Errorf("expected 'invalid cancellation date' error, got %v", err)
		}
	})
}
