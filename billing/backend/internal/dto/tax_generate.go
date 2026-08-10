package dto

import (
	"billing/internal/port"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TaxGenerateRequest represents the request data for sending emissions.
type TaxGenerateRequest struct {
	Vendor           string `json:"vendor"`
	VendorID         int64  `json:"-"`
	InvoiceStartDate string `json:"invoice_start_date"`
	InvoiceEndDate   string `json:"invoice_end_date"`
	EmissionDate     string `json:"emission_date"`
}

// TaxGenerateResponse represents the response data after sending emissions.
type TaxGenerateResponse struct {
	ResponseBase
	EmissionID       int64   `json:"emission_id"`
	EmissionQuantity int     `json:"emission_quantity"`
	EmissionAmount   float64 `json:"emission_amount"`
	DocumentBase64   string  `json:"document_base64"`
	DocumentName     string  `json:"document_name"`
}

// NewTaxGenerateResponse creates a new instance of TaxGenerateResponse.
func NewTaxGenerateResponse(httpCode int, status, message string, emissionID int64,
	emissionQuantity int, emissionAmount float64, documentBase64, documentName string) TaxGenerateResponse {
	return TaxGenerateResponse{
		ResponseBase:     NewResponseBase(httpCode, status, message),
		EmissionID:       emissionID,
		EmissionQuantity: emissionQuantity,
		EmissionAmount:   emissionAmount,
		DocumentBase64:   documentBase64,
		DocumentName:     documentName,
	}
}

// Validate checks if the TaxGenerateRequest has all required fields and valid data.
func (r *TaxGenerateRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDates(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return err
	}

	return nil
}

// valdateVendor checks if the vendor exists in the repository.
func (r *TaxGenerateRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return fmt.Errorf("vendor is required")
	}
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return fmt.Errorf("failed to validate vendor: %v", err)
	}
	if vendor == nil {
		return fmt.Errorf("vendor '%s' does not exist", r.Vendor)
	}
	r.VendorID = vendor.ID
	return nil
}

// validateDates checks if the provided dates are valid and in the correct format.
func (r *TaxGenerateRequest) validateDates() error {
	errs := make([]error, 0)
	if r.InvoiceStartDate == "" {
		errs = append(errs, fmt.Errorf("invoice_start_date is required"))
	} else {
		if _, err := time.Parse("2006-01-02", r.InvoiceStartDate); err != nil {
			errs = append(errs, fmt.Errorf("invalid invoice_start_date format: %v", err))
		}
	}
	if r.InvoiceEndDate == "" {
		errs = append(errs, fmt.Errorf("invoice_end_date is required"))
	} else {
		if _, err := time.Parse("2006-01-02", r.InvoiceEndDate); err != nil {
			errs = append(errs, fmt.Errorf("invalid invoice_end_date format: %v", err))
		} else {
			startDate, _ := time.Parse("2006-01-02", r.InvoiceStartDate)
			endDate, _ := time.Parse("2006-01-02", r.InvoiceEndDate)
			if endDate.Before(startDate) {
				errM := fmt.Errorf("invoice_end_date cannot be before invoice_start_date")
				errs = append(errs, errM)
			}
		}
	}

	if r.EmissionDate == "" {
		errs = append(errs, fmt.Errorf("emission_date is required"))
	} else {
		if _, err := time.Parse("2006-01-02", r.EmissionDate); err != nil {
			errs = append(errs, fmt.Errorf("invalid emission_date format: %v", err))
		} else {
			emissionDate, _ := time.Parse("2006-01-02", r.EmissionDate)
			endDate, _ := time.Parse("2006-01-02", r.InvoiceEndDate)
			if emissionDate.Before(endDate) {
				errM := fmt.Errorf("emission_date cannot be before invoice_end_date")
				errs = append(errs, errM)
			}
		}
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// Reset clears the fields of the TaxGenerateRequest.
func (r *TaxGenerateRequest) Reset() {
	r.Vendor = ""
	r.VendorID = 0
	r.InvoiceStartDate = ""
	r.InvoiceEndDate = ""
	r.EmissionDate = ""
}
