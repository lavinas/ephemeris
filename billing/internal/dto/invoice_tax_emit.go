package dto

import (
	"errors"
	"fmt"
	"strings"

	"billing/internal/domain"
	"billing/internal/port"
)

// InvoiceTaxEmitRequest represents the DTO for emitting a single NFS-e for an invoice.
type InvoiceTaxEmitRequest struct {
	Vendor    string          `json:"vendor"`
	vendorID  int64           `json:"-" validate:"-"`
	ID        int64           `json:"id"`
	InvoiceID int64           `json:"invoice_id,omitempty"`
	invoice   *domain.Invoice `json:"-" validate:"-"`
	vendor    *domain.Vendor  `json:"-" validate:"-"`
}

// InvoiceTaxEmitResponse represents the response after emitting a single NFS-e.
type InvoiceTaxEmitResponse struct {
	ResponseBase
	NFENumber       *int64  `json:"nfe_number,omitempty"`
	NFEVerification *string `json:"nfe_verification,omitempty"`
	NFEDatetime     *string `json:"nfe_datetime,omitempty"`
	RPSNumber       int64   `json:"rps_number,omitempty"`
}

// NewInvoiceTaxEmitResponse creates a new instance of InvoiceTaxEmitResponse.
func NewInvoiceTaxEmitResponse(httpCode int, status, message string, nfeNumber *int64, nfeVerif *string, nfeDate *string, rpsNumber int64) InvoiceTaxEmitResponse {
	return InvoiceTaxEmitResponse{
		ResponseBase:    NewResponseBase(httpCode, status, message),
		NFENumber:       nfeNumber,
		NFEVerification: nfeVerif,
		NFEDatetime:     nfeDate,
		RPSNumber:       rpsNumber,
	}
}

// Validate validates the InvoiceTaxEmitRequest against repository and domain rules.
func (r *InvoiceTaxEmitRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateInvoice(repo); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *InvoiceTaxEmitRequest) validateVendor(repo port.Repository) error {
	if strings.TrimSpace(r.Vendor) == "" {
		return errors.New("vendor is required")
	}
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return err
	}
	if vendor == nil {
		return fmt.Errorf("vendor '%s' not found", r.Vendor)
	}
	r.vendorID = vendor.ID
	r.vendor = vendor
	return nil
}

func (r *InvoiceTaxEmitRequest) validateInvoice(repo port.Repository) error {
	if r.ID <= 0 && r.InvoiceID > 0 {
		r.ID = r.InvoiceID
	}
	if r.ID <= 0 {
		return errors.New("id must be a positive integer")
	}
	invoice, err := repo.GetInvoice(r.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch invoice: %v", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice with ID %d not found for vendor '%s'", r.ID, r.Vendor)
	}
	if r.vendorID != 0 && invoice.Customer.VendorID != r.vendorID {
		return fmt.Errorf("invoice with ID %d does not belong to vendor '%s'", r.ID, r.Vendor)
	}
	if invoice.CancellationDate != nil {
		return errors.New("cannot emit tax note for a canceled invoice")
	}
	if invoice.TaxDate != nil {
		return errors.New("tax note has already been emitted for this invoice")
	}
	if invoice.PaymentDate == nil {
		return errors.New("invoice must be paid before emitting tax note")
	}
	if invoice.Customer.Document == nil || strings.TrimSpace(*invoice.Customer.Document) == "" {
		return errors.New("customer document (CPF/CNPJ) is required to emit tax note")
	}
	if r.vendor != nil && (invoice.Vendor.Document == "" || invoice.Vendor.TaxDocument == "") {
		invoice.Vendor = *r.vendor
		invoice.VendorID = r.vendor.ID
	}
	r.invoice = invoice
	return nil
}

// GetInvoice returns the validated domain invoice.
func (r *InvoiceTaxEmitRequest) GetInvoice() *domain.Invoice {
	return r.invoice
}

// GetVendor returns the validated domain vendor.
func (r *InvoiceTaxEmitRequest) GetVendor() *domain.Vendor {
	return r.vendor
}

// Reset resets all fields in InvoiceTaxEmitRequest.
func (r *InvoiceTaxEmitRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.ID = 0
	r.InvoiceID = 0
	r.invoice = nil
	r.vendor = nil
}
