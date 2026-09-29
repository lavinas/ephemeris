package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/port"
)

// InvoicePaymentRequest represents the data transfer object for registering an invoice payment.
type InvoicePaymentRequest struct {
	Vendor            string          `json:"vendor"`
	vendorID          int64           `json:"-" validate:"-"`
	ID                int64           `json:"id"`
	InvoiceID         int64           `json:"invoice_id,omitempty"`
	PaymentDate       string          `json:"payment_date"`
	Payment           string          `json:"payment,omitempty"`
	invoice           *domain.Invoice `json:"-" validate:"-"`
	parsedPaymentDate time.Time       `json:"-" validate:"-"`
}

// InvoicePaymentResponse represents the response after processing invoice payment.
type InvoicePaymentResponse struct {
	ResponseBase
}

// NewInvoicePaymentResponse creates a new instance of InvoicePaymentResponse.
func NewInvoicePaymentResponse(httpCode int, status, message string) InvoicePaymentResponse {
	return InvoicePaymentResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate validates the InvoicePaymentRequest fields against business rules and repository state.
func (r *InvoicePaymentRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateInvoice(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePaymentDate(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *InvoicePaymentRequest) validateVendor(repo port.Repository) error {
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
	return nil
}

func (r *InvoicePaymentRequest) validateInvoice(repo port.Repository) error {
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
		return errors.New("cannot process payment for a canceled invoice")
	}
	if invoice.PaymentDate != nil {
		return errors.New("invoice is already paid")
	}
	if invoice.Customer.Email == nil || strings.TrimSpace(*invoice.Customer.Email) == "" {
		return errors.New("customer has no email address configured to receive the receipt")
	}
	r.invoice = invoice
	return nil
}

func (r *InvoicePaymentRequest) validatePaymentDate() error {
	if strings.TrimSpace(r.PaymentDate) == "" && strings.TrimSpace(r.Payment) != "" {
		r.PaymentDate = strings.TrimSpace(r.Payment)
	}
	if strings.TrimSpace(r.PaymentDate) == "" {
		return errors.New("payment_date is required")
	}
	dt, err := time.Parse("2006-01-02", strings.TrimSpace(r.PaymentDate))
	if err != nil {
		return fmt.Errorf("invalid payment_date format, expected YYYY-MM-DD: %v", err)
	}
	r.parsedPaymentDate = dt
	return nil
}

// GetInvoice returns the validated domain invoice.
func (r *InvoicePaymentRequest) GetInvoice() *domain.Invoice {
	return r.invoice
}

// GetPaymentDate returns the parsed payment date.
func (r *InvoicePaymentRequest) GetPaymentDate() time.Time {
	return r.parsedPaymentDate
}

// Reset resets all fields in InvoicePaymentRequest.
func (r *InvoicePaymentRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.ID = 0
	r.InvoiceID = 0
	r.PaymentDate = ""
	r.Payment = ""
	r.invoice = nil
	r.parsedPaymentDate = time.Time{}
}
