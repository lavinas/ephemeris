package dto

import (
	"errors"
	"strings"

	"billing/internal/port"
)

// BillGetRequest represents a request to get an invoice bill for a customer.
type BillGetRequest struct {
	DocumentType int    `json:"document_type"` // 1 - bill, 2 - qrcode, 3 - payload
	Vendor       string `json:"vendor"`
	vendorID     int64  `json:"-" validate:"-"`
	InvoiceID    int64  `json:"invoice_id"`
}

// BillGetResponse represents the response after retrieving an invoice bill for a customer.
type BillGetResponse struct {
	ResponseBase
	DocumentType   int    `json:"document_type"` // 1 - bill, 2 - qrcode, 3 - payload
	DocumentBase64 string `json:"document_base64"`
}

// NewBillGetResponse creates a new instance of BillGetResponse with the provided parameters.
func NewBillGetResponse(statusCode int, statusMessage string, errorMessage string, documentType int, documentBase64 string) *BillGetResponse {
	return &BillGetResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		DocumentType:   documentType,
		DocumentBase64: documentBase64,
	}
}

// Validate checks if the BillGetRequest has valid data.
func (r *BillGetRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateDocumentType(); err != nil {
		errs = append(errs, err)
	}
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

// validateDocumentType checks if the DocumentType is valid (1, 2, or 3).
func (r *BillGetRequest) validateDocumentType() error {
	if r.DocumentType < 1 || r.DocumentType > 3 {
		return errors.New("invalid document type: must be 1 (bill), 2 (qrcode), or 3 (pix code)")
	}
	return nil
}

// validateVendor checks if the vendor exists in the repository and sets the vendorID in the request.
func (r *BillGetRequest) validateVendor(repo port.Repository) error {
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return err
	}
	r.vendorID = vendor.ID
	return nil
}

// validateInvoiceID checks if the InvoiceID is valid (greater than 0).
func (r *BillGetRequest) validateInvoiceID(repo port.Repository) error {
	invoice, err := repo.GetInvoice(r.InvoiceID)
	if err != nil {
		return err
	}
	if invoice == nil {
		return errors.New("invoice not found")
	}
	if invoice.Customer.VendorID != r.vendorID {
		return errors.New("invoice does not belong to the specified vendor")
	}
	return nil
}

// Reset resets the BillGetRequest fields to their zero values.
func (r *BillGetRequest) Reset() {
	r.DocumentType = 0
	r.Vendor = ""
	r.InvoiceID = 0
}
