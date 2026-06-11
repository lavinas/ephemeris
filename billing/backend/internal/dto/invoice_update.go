package dto

import (
	"errors"
	"fmt"
	"time"
	"strings"

	"billing/internal/port"
	"billing/internal/domain"
)

// InvoiceUpdateRequest represents the data transfer object for updating an existing invoice.
type InvoiceUpdateRequest struct {
	Vendor		     string               `json:"vendor" validate:"required"`
	vendorID	     int64                `json:"-" validate:"-"`
	ID			     int64                `json:"id" validate:"required"`
	InvoiceDate      *string              `json:"invoicing,omitempty"`
	DueDate          *string              `json:"due,omitempty"`
	PaymentDate      *string              `json:"payment,omitempty"`
	EmailSentDate    *string              `json:"email_sent,omitempty"`
	WhatsappSentDate *string              `json:"whatsapp_sent,omitempty"`
	TaxDate          *string              `json:"tax,omitempty"`
	CancellationDate *string              `json:"cancellation,omitempty"`
	invoice          *domain.Invoice      `json:"-" validate:"-"`
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
	if err := r.validateEmailSentDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsappSentDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateTaxDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateCancellationDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateOneOfDates(); err != nil {
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
	if invoice.Customer.VendorID != r.vendorID {
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

// validateEmailSentDate checks if the provided email sent date is valid.
func (r *InvoiceUpdateRequest) validateEmailSentDate() error {
	if r.EmailSentDate == nil {
		return nil // No update for email sent date
	}
	_, err := time.Parse("2006-01-02", *r.EmailSentDate)
	if err != nil {
		return fmt.Errorf("invalid email sent date: %v", err)
	}
	return nil
}

// validateWhatsappSentDate checks if the provided WhatsApp sent date is valid.
func (r *InvoiceUpdateRequest) validateWhatsappSentDate() error {
	if r.WhatsappSentDate == nil {
		return nil // No update for WhatsApp sent date
	}
	_, err := time.Parse("2006-01-02", *r.WhatsappSentDate)
	if err != nil {
		return fmt.Errorf("invalid WhatsApp sent date: %v", err)
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

// validateOneOfDates checks if at least one of the date fields is provided for update.
func (r *InvoiceUpdateRequest) validateOneOfDates() error {
	if r.InvoiceDate == nil && r.DueDate == nil && r.PaymentDate == nil &&
		r.EmailSentDate == nil && r.WhatsappSentDate == nil &&
		r.TaxDate == nil && r.CancellationDate == nil {
		return fmt.Errorf("at least one date field must be provided for update")
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
	if r.EmailSentDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.EmailSentDate)
		r.invoice.EmailSentDate = &dt
	}
	if r.WhatsappSentDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.WhatsappSentDate)
		r.invoice.WhatsappSentDate = &dt
	}
	if r.TaxDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.TaxDate)
		r.invoice.TaxDate = &dt
	}
	if r.CancellationDate != nil {
		dt, _ := time.Parse("2006-01-02", *r.CancellationDate)
		r.invoice.CancellationDate = &dt
	}
	r.invoice.UpdatedAt = time.Now()
	return r.invoice
}

// ResetDomain resets the domain object to nil after processing the update request.
func (r *InvoiceUpdateRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.ID = 0
	r.InvoiceDate = nil
	r.DueDate = nil
	r.PaymentDate = nil
	r.EmailSentDate = nil
	r.WhatsappSentDate = nil
	r.TaxDate = nil
	r.CancellationDate = nil
	r.invoice = nil
}
