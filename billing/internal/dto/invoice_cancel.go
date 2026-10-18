package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/port"
)

// InvoiceCancelRequest represents the data transfer object for cancelling an invoice.
type InvoiceCancelRequest struct {
	Vendor           string          `json:"vendor"`
	vendorID         int64           `json:"-" validate:"-"`
	ID               int64           `json:"id,omitempty"`
	InvoiceID        int64           `json:"invoice_id,omitempty"`
	ExternalCode     *string         `json:"external_code,omitempty"`
	CancellationDate *string         `json:"cancellation,omitempty"`
	Cancellation     *string         `json:"cancellation_date,omitempty"`
	invoice          *domain.Invoice `json:"-" validate:"-"`
	parsedDate       time.Time       `json:"-" validate:"-"`
}

// InvoiceCancelResponse represents the response data transfer object after cancelling an invoice.
type InvoiceCancelResponse struct {
	ResponseBase
	ID *int64 `json:"id,omitempty"`
}

// NewInvoiceCancelResponse creates a new instance of InvoiceCancelResponse.
func NewInvoiceCancelResponse(httpCode int, status, message string, id *int64) InvoiceCancelResponse {
	return InvoiceCancelResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		ID:           id,
	}
}

// Validate validates the InvoiceCancelRequest fields against business rules and repository state.
func (r *InvoiceCancelRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateInvoice(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateCancellationDate(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *InvoiceCancelRequest) validateVendor(repo port.Repository) error {
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

func (r *InvoiceCancelRequest) validateInvoice(repo port.Repository) error {
	if r.ID <= 0 && r.InvoiceID > 0 {
		r.ID = r.InvoiceID
	}

	var invoice *domain.Invoice
	var err error

	if r.ID > 0 {
		invoice, err = repo.GetInvoice(r.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch invoice: %v", err)
		}
		if invoice == nil {
			return fmt.Errorf("invoice with ID %d not found", r.ID)
		}
	} else if r.ExternalCode != nil && strings.TrimSpace(*r.ExternalCode) != "" {
		trimmedCode := strings.TrimSpace(*r.ExternalCode)
		if r.vendorID != 0 {
			invoice, err = repo.GetInvoiceByExternalCode(r.vendorID, trimmedCode)
			if err != nil {
				return fmt.Errorf("failed to fetch invoice by external_code: %v", err)
			}
			if invoice == nil {
				return fmt.Errorf("invoice with external_code '%s' not found for vendor '%s'", trimmedCode, r.Vendor)
			}
		} else {
			return errors.New("cannot lookup external_code without valid vendor")
		}
	} else {
		return errors.New("id or external_code is required")
	}

	invoiceVendorID := invoice.VendorID
	if invoiceVendorID == 0 {
		invoiceVendorID = invoice.Customer.VendorID
	}
	if r.vendorID != 0 && invoiceVendorID != r.vendorID {
		return fmt.Errorf("invoice with ID %d does not belong to vendor '%s'", invoice.ID, r.Vendor)
	}

	if invoice.CancellationDate != nil {
		return fmt.Errorf("invoice with ID %d is already cancelled", invoice.ID)
	}
	if invoice.PaymentDate != nil {
		return fmt.Errorf("invoice with ID %d cannot be cancelled: already paid", invoice.ID)
	}
	if invoice.TaxDate != nil {
		return fmt.Errorf("invoice with ID %d cannot be cancelled: tax document already issued", invoice.ID)
	}

	r.invoice = invoice
	return nil
}

func (r *InvoiceCancelRequest) validateCancellationDate() error {
	dateStr := ""
	if r.CancellationDate != nil && strings.TrimSpace(*r.CancellationDate) != "" {
		dateStr = strings.TrimSpace(*r.CancellationDate)
	} else if r.Cancellation != nil && strings.TrimSpace(*r.Cancellation) != "" {
		dateStr = strings.TrimSpace(*r.Cancellation)
	}

	if dateStr == "" {
		now := time.Now()
		r.parsedDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return nil
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid cancellation date: %v", err)
	}
	r.parsedDate = parsed
	return nil
}

// GetDomain updates the domain invoice with the cancellation date and returns it.
func (r *InvoiceCancelRequest) GetDomain() *domain.Invoice {
	if r.invoice == nil {
		return nil
	}
	r.invoice.CancellationDate = &r.parsedDate
	if r.invoice.ExternalCode != nil && *r.invoice.ExternalCode != "" {
		newCode := fmt.Sprintf("%s-CAN-%d", *r.invoice.ExternalCode, time.Now().Unix())
		if len(newCode) > 50 {
			newCode = newCode[:50]
		}
		r.invoice.ExternalCode = &newCode
	}
	return r.invoice
}

// Reset clears the request data.
func (r *InvoiceCancelRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.ID = 0
	r.InvoiceID = 0
	r.ExternalCode = nil
	r.CancellationDate = nil
	r.Cancellation = nil
	r.invoice = nil
	r.parsedDate = time.Time{}
}
