package dto

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"billing/internal/port"
)

// BillRequest represents a request to send a receipt to a customer.
type BillRequest struct {
	Vendor    string `json:"vendor"`
	vendorID  int64  `json:"-" validate:"-"`
	InvoiceID int64  `json:"invoice_id"`
	Doc       int    `json:"doc"`    // 0 - Invoice, 1 - Receipt
	Action    int    `json:"action"` // 0 - send email, 1 - resend email, 2 - get pdf base64
	Email     string `json:"email"`  // optional, if not provided, the email associated with the invoice will be used. Cannot be used with Action 0 (send email)
}

// BillResponse represents the response after sending a bill to a customer.
type BillResponse struct {
	ResponseBase
	DocumentBase64 *string `json:"document_base64,omitempty"`
	DocumentName   *string `json:"document_name,omitempty"`
}

// NewBillResponse creates a new instance of BillResponse with the provided parameters.
func NewBillResponse(statusCode int, statusMessage string, errorMessage string, documentBase64 *string, documentName *string) *BillResponse {
	return &BillResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		DocumentBase64: documentBase64,
		DocumentName:   documentName,
	}
}

// Validate checks if the BillRequest has valid data.
func (r *BillRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAction(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDoc(); err != nil {
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

// validateVendor checks if the Vendor is valid and exists in the repository.
func (r *BillRequest) validateVendor(repo port.Repository) error {
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
func (r *BillRequest) validateInvoice(repo port.Repository) error {
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
	if r.Doc == 1 && invoice.PaymentDate == nil {
		return errors.New("cannot send receipt for an unpaid invoice")
	}
	if invoice.CancellationDate != nil {
		return errors.New("cannot send receipt for a canceled invoice")
	}
	if r.Action == 0 {
		if r.Doc == 0 && invoice.EmailSentDate != nil {
			return errors.New("email has already been sent for this invoice")
		}
		if r.Doc == 1 && invoice.EmailReceiptDate != nil {
			return errors.New("email has already been sent for this receipt")
		}
	}

	return nil
}

// Reset clears the fields of the BillRequest.
func (r *BillRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.InvoiceID = 0
	r.Email = ""
}

// validateEmail checks if the provided email is valid and not already in use.
func (r *BillRequest) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	if r.Action == 0 {
		return errors.New("send mail action requires no email address")
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// validateAction checks if the Action is valid.
func (r *BillRequest) validateAction() error {
	if r.Action < 0 || r.Action > 2 {
		return errors.New("action must be 0 (send email), 1 (resend email), or 2 (get pdf base64)")
	}
	return nil
}

// validateDoc checks if the Doc field is valid.
func (r *BillRequest) validateDoc() error {
	if r.Doc < 0 || r.Doc > 1 {
		return errors.New("doc must be 0 (Invoice) or 1 (Receipt)")
	}
	return nil
}
