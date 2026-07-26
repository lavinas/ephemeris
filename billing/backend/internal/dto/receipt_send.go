package dto

import (
	"errors"
	"strings"

	"billing/internal/port"
)

// ReceiptSendRequest represents a request to send a receipt to a customer.
type ReceiptSendRequest struct {
	Vendor    string `json:"vendor"`
	vendorID  int64  `json:"-" validate:"-"`
	InvoiceID int64  `json:"invoice_id"`
}

// ReceiptSendResponse represents the response after sending a receipt to a customer.
type ReceiptSendResponse struct {
	ResponseBase
}

// NewReceiptSendResponse creates a new instance of ReceiptSendResponse with the provided parameters.
func NewReceiptSendResponse(statusCode int, statusMessage string, errorMessage string) *ReceiptSendResponse {
	return &ReceiptSendResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
	}
}

// Validate checks if the ReceiptSendRequest has valid data.
func (r *ReceiptSendRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateInvoiceID(repo); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateVendor checks if the Vendor is valid and exists in the repository.
func (r *ReceiptSendRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return errors.New("vendor is required")
	}
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return err
	}
	r.vendorID = vendor.ID
	return nil
}

// validateInvoiceID checks if the InvoiceID is valid and exists in the repository.
func (r *ReceiptSendRequest) validateInvoiceID(repo port.Repository) error {
	if r.InvoiceID <= 0 {
		return errors.New("invoice_id must be a positive integer")
	}
	invoice, err := repo.GetInvoice(r.InvoiceID)
	if err != nil {
		return err
	}
	if invoice.Customer.VendorID != r.vendorID {
		return errors.New("invoice does not belong to the specified vendor")
	}
	return nil
}

// Reset clears the fields of the ReceiptSendRequest.
func (r *ReceiptSendRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.InvoiceID = 0
}
