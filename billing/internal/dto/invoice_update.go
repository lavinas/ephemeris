package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/port"
)

// InvoiceUpdateRequest represents the data transfer object for updating an existing invoice.
type InvoiceUpdateRequest struct {
	Vendor           string          `json:"vendor" validate:"required"`
	vendorID         int64           `json:"-" validate:"-"`
	ID               int64           `json:"id" validate:"required"`
	InvoiceDate      *string         `json:"invoicing,omitempty"`
	DueDate          *string         `json:"due,omitempty"`
	PaymentDate      *string         `json:"payment,omitempty"`
	SentDate         *string         `json:"sent,omitempty"`
	EmailSentDate    *string         `json:"email_sent,omitempty"`
	ReceiptDate      *string         `json:"receipt,omitempty"`
	EmailReceiptDate *string         `json:"email_receipt,omitempty"`
	TaxDate          *string         `json:"tax,omitempty"`
	CancellationDate *string         `json:"cancellation,omitempty"`
	Notes            *string         `json:"notes,omitempty"`
	ExternalCode     *string         `json:"external_code,omitempty"`
	invoice          *domain.Invoice `json:"-" validate:"-"`
}

// InvoiceUpdateResponse represents the data transfer object for the response of updating an existing invoice.
type InvoiceUpdateResponse struct {
	ResponseBase
}

// NewInvoiceUpdateResponse creates a new instance of InvoiceUpdateResponse
func NewInvoiceUpdateResponse(httpCode int, status, message string) InvoiceUpdateResponse {
	return InvoiceUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate validates the InvoiceUpdateRequest fields using the provided validator.
func (r *InvoiceUpdateRequest) Validate(repo port.Repository) error {
	if r.SentDate == nil && r.EmailSentDate != nil {
		r.SentDate = r.EmailSentDate
	}
	if r.ReceiptDate == nil && r.EmailReceiptDate != nil {
		r.ReceiptDate = r.EmailReceiptDate
	}

	errs := make([]error, 0)
	if err := r.ValidateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateID(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateInvoiceDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDueDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePaymentDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSentDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateReceiptDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateTaxDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateCancellationDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateExternalCode(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateOneOfFields(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// ValidateVendor checks if the provided vendor is valid and sets the vendorID.
func (r *InvoiceUpdateRequest) ValidateVendor(repo port.Repository) error {
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return err
	}
	if vendor == nil {
		return fmt.Errorf("vendor '%s' not found", r.Vendor)
	}
	r.vendorID = vendor.ID
	return nil
}

// validateID checks if the provided ID is valid.
func (r *InvoiceUpdateRequest) validateID(repo port.Repository) error {
	if r.ID <= 0 {
		return fmt.Errorf("invalid invoice ID: %d", r.ID)
	}
	invoice, err := repo.GetInvoice(r.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch invoice: %v", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice with ID %d not found for vendor '%s'", r.ID, r.Vendor)
	}
	invoiceVendorID := invoice.VendorID
	if invoiceVendorID == 0 {
		invoiceVendorID = invoice.Customer.VendorID
	}
	if invoiceVendorID != r.vendorID {
		return fmt.Errorf("invoice with ID %d does not belong to vendor '%s'", r.ID, r.Vendor)
	}
	r.invoice = invoice
	return nil
}

// validateInvoiceDate checks if the provided invoice date is valid.
func (r *InvoiceUpdateRequest) validateInvoiceDate() error {
	if r.InvoiceDate == nil {
		return nil // No update for invoice date
	}
	_, err := time.Parse("2006-01-02", *r.InvoiceDate)
	if err != nil {
		return fmt.Errorf("invalid invoice date: %v", err)
	}
	return nil
}

// validateDueDate checks if the provided due date is valid.
func (r *InvoiceUpdateRequest) validateDueDate() error {
	if r.DueDate == nil {
		return nil // No update for due date
	}
	_, err := time.Parse("2006-01-02", *r.DueDate)
	if err != nil {
		return fmt.Errorf("invalid due date: %v", err)
	}
	return nil
}

// validatePaymentDate checks if the provided payment date is valid.
func (r *InvoiceUpdateRequest) validatePaymentDate() error {
	if r.PaymentDate == nil {
		return nil // No update for payment date
	}
	_, err := time.Parse("2006-01-02", *r.PaymentDate)
	if err != nil {
		return fmt.Errorf("invalid payment date: %v", err)
	}
	return nil
}

// validateSentDate checks if the provided sent date is valid.
func (r *InvoiceUpdateRequest) validateSentDate() error {
	if r.SentDate == nil {
		return nil // No update for sent date
	}
	_, err := time.Parse("2006-01-02", *r.SentDate)
	if err != nil {
		return fmt.Errorf("invalid sent date: %v", err)
	}
	return nil
}

// validateReceiptDate checks if the provided receipt date is valid.
func (r *InvoiceUpdateRequest) validateReceiptDate() error {
	if r.ReceiptDate == nil {
		return nil
	}
	_, err := time.Parse("2006-01-02", *r.ReceiptDate)
	if err != nil {
		return fmt.Errorf("invalid receipt date: %v", err)
	}
	return nil
}

// validateTaxDate checks if the provided tax date is valid.
func (r *InvoiceUpdateRequest) validateTaxDate() error {
	if r.TaxDate == nil {
		return nil // No update for tax date
	}
	_, err := time.Parse("2006-01-02", *r.TaxDate)
	if err != nil {
		return fmt.Errorf("invalid tax date: %v", err)
	}
	return nil
}

// validateCancellationDate checks if the provided cancellation date is valid.
func (r *InvoiceUpdateRequest) validateCancellationDate() error {
	if r.CancellationDate == nil {
		return nil // No update for cancellation date
	}
	_, err := time.Parse("2006-01-02", *r.CancellationDate)
	if err != nil {
		return fmt.Errorf("invalid cancellation date: %v", err)
	}
	return nil
}

// validateExternalCode validates format and checks duplicates for vendor.
func (r *InvoiceUpdateRequest) validateExternalCode(repo port.Repository) error {
	if r.ExternalCode == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*r.ExternalCode)
	if trimmed == "" {
		r.ExternalCode = nil
		return nil
	}
	r.ExternalCode = &trimmed
	if len(trimmed) > 50 {
		return fmt.Errorf("external_code cannot exceed 50 characters")
	}
	if r.vendorID != 0 {
		existing, err := repo.GetInvoiceByExternalCode(r.vendorID, trimmed)
		if err != nil {
			return fmt.Errorf("error verifying external_code: %v", err)
		}
		if existing != nil && existing.ID != r.ID {
			return fmt.Errorf("duplicate external_code '%s' for vendor", trimmed)
		}
	}
	return nil
}

// validateOneOfFields checks if at least one field is provided for update.
func (r *InvoiceUpdateRequest) validateOneOfFields() error {
	if r.InvoiceDate == nil && r.DueDate == nil && r.PaymentDate == nil &&
		r.SentDate == nil && r.ReceiptDate == nil &&
		r.TaxDate == nil && r.CancellationDate == nil &&
		r.Notes == nil && r.ExternalCode == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// GetDomain constructs a map of the fields to be updated based on the non-empty fields
func (r *InvoiceUpdateRequest) GetDomain() interface{} {
	if r.invoice == nil {
		return nil
	}
	if r.InvoiceDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.InvoiceDate)
		r.invoice.InvoiceDate = dt
	}
	if r.DueDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.DueDate)
		r.invoice.DueDate = dt
	}
	if r.PaymentDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.PaymentDate)
		r.invoice.PaymentDate = &dt
	}
	if r.SentDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.SentDate)
		r.invoice.SentDate = &dt
	}
	if r.ReceiptDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.ReceiptDate)
		r.invoice.ReceiptDate = &dt
	}
	if r.TaxDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.TaxDate)
		r.invoice.TaxDate = &dt
	}
	if r.CancellationDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.CancellationDate)
		r.invoice.CancellationDate = &dt
	}
	if r.Notes != nil {
		r.invoice.Notes = r.Notes
	}
	if r.ExternalCode != nil {
		r.invoice.ExternalCode = r.ExternalCode
	}
	r.invoice.UpdatedAt = time.Now()
	return r.invoice
}

// Reset resets the update request fields.
func (r *InvoiceUpdateRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.ID = 0
	r.InvoiceDate = nil
	r.DueDate = nil
	r.PaymentDate = nil
	r.SentDate = nil
	r.EmailSentDate = nil
	r.ReceiptDate = nil
	r.EmailReceiptDate = nil
	r.TaxDate = nil
	r.CancellationDate = nil
	r.Notes = nil
	r.ExternalCode = nil
	r.invoice = nil
}
