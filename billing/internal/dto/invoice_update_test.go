package dto

import (
	"strings"
	"testing"

	"billing/internal/domain"
)

func TestInvoiceUpdate_ExternalCode_Validation(t *testing.T) {
	vendor := &domain.Vendor{ID: 10, Nickname: "estudio_amelia"}
	customer := &domain.Customer{ID: 20, VendorID: 10, Nickname: "cliente_1"}
	existingCode := "EXT-EXISTING"
	repo := &invoiceCreateMockRepo{
		vendor:   vendor,
		customer: customer,
		invoices: map[string]*domain.Invoice{
			existingCode: {ID: 99, VendorID: 10, ExternalCode: &existingCode},
		},
	}

	// Case 1: Updating invoice 1 to duplicate external_code of invoice 99 -> Should fail
	code := "EXT-EXISTING"
	req := &InvoiceUpdateRequest{
		ID:           1,
		Vendor:       "estudio_amelia",
		ExternalCode: &code,
	}

	err := req.Validate(repo)
	if err == nil {
		t.Fatalf("expected error for duplicate external_code, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate external_code 'EXT-EXISTING' for vendor") {
		t.Errorf("unexpected error message: %v", err)
	}

	// Case 2: Updating invoice 99 to its own external_code -> Should succeed
	reqOwn := &InvoiceUpdateRequest{
		ID:           99,
		Vendor:       "estudio_amelia",
		ExternalCode: &code,
	}
	if err := reqOwn.Validate(repo); err != nil {
		t.Fatalf("unexpected error when keeping own external_code: %v", err)
	}

	// Case 3: Updating invoice 1 to a unique external_code -> Should succeed
	newCode := "EXT-NEW"
	reqNew := &InvoiceUpdateRequest{
		ID:           1,
		Vendor:       "estudio_amelia",
		ExternalCode: &newCode,
	}
	if err := reqNew.Validate(repo); err != nil {
		t.Fatalf("unexpected error for new external_code: %v", err)
	}
}

func TestInvoiceUpdate_DateValidation(t *testing.T) {
	vendor := &domain.Vendor{ID: 10, Nickname: "estudio_amelia"}
	repo := &invoiceCreateMockRepo{vendor: vendor}

	invalidDate := "invalid-date"
	req := &InvoiceUpdateRequest{
		ID:       1,
		Vendor:   "estudio_amelia",
		SentDate: &invalidDate,
	}
	err := req.Validate(repo)
	if err == nil {
		t.Fatalf("expected error for invalid SentDate format, got nil")
	}

	validDate := "2026-10-15"
	req.SentDate = &validDate
	if err := req.Validate(repo); err != nil {
		t.Fatalf("unexpected error for valid SentDate: %v", err)
	}
}
