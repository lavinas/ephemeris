package dto

import (
	"billing/internal/port"
	"errors"
	"fmt"
	"strings"
	"time"
)

// TaxSendRequest represents the request data for sending emissions.
type TaxSendRequest struct {
	Vendor           string `json:"vendor"`
	VendorID         int64  `json:"-"`
	InvoiceStartDate string `json:"invoice_start_date"`
	InvoiceEndDate   string `json:"invoice_end_date"`
	EmissionDate     string `json:"emission_date"`
}

// TaxSendResponse represents the response data after sending emissions.
type TaxSendResponse struct {
	ResponseBase
	EmissionID       int64   `json:"emission_id"`
	EmissionQuantity int     `json:"emission_quantity"`
	EmissionAmount   float64 `json:"emission_amount"`
}

// NewTaxSendResponse creates a new instance of TaxSendResponse.
func NewTaxSendResponse(httpCode int, status, message string, emissionID int64,
	emissionQuantity int, emissionAmount float64) TaxSendResponse {
	return TaxSendResponse{
		ResponseBase:     NewResponseBase(httpCode, status, message),
		EmissionID:       emissionID,
		EmissionQuantity: emissionQuantity,
		EmissionAmount:   emissionAmount,
	}
}

// Validate checks if the TaxSendRequest has all required fields and valid data.
func (r *TaxSendRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.valdateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDates(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDuplication(repo); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return err
	}

	return nil
}

// valdateVendor checks if the vendor exists in the repository.
func (r *TaxSendRequest) valdateVendor(repo port.Repository) error {
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
func (r *TaxSendRequest) validateDates() error {
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

// validateDuplication checks if an emission for the given vendor and period
// already exists in the repository.
func (r *TaxSendRequest) validateDuplication(repo port.Repository) error {
	// Implement logic to check for existing emissions in the repository.
	// This is a placeholder implementation; replace with actual database query.
	startDate, _ := time.Parse("2006-01-02", r.InvoiceStartDate)
	endDate, _ := time.Parse("2006-01-02", r.InvoiceEndDate)
	count, err := repo.GetEmissionsCount(r.VendorID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to check for existing emissions: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("an emission for vendor '%s' already exists", r.Vendor)
	}
	return nil
}

// Reset clears the fields of the TaxSendRequest.
func (r *TaxSendRequest) Reset() {
	r.Vendor = ""
	r.VendorID = 0
	r.InvoiceStartDate = ""
	r.InvoiceEndDate = ""
	r.EmissionDate = ""
}
