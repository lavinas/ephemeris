package dto

import (
	"billing/internal/port"
	"fmt"
	"errors"
	"time"
)

// EmissionSendRequest represents the request data for sending emissions.
type EmissionSendRequest struct {
	Vendor   string `json:"vendor"`
	vendorID         int64  `json:"-"`
	InvoiceStartDate string `json:"invoice_start_date"`
	InvoiceEndDate   string `json:"invoice_end_date"`
	EmissionDate     string `json:"emission_date"`
}

// EmissionSendResponse represents the response data after sending emissions.
type EmissionSendResponse struct {
	ResponseBase
	EmissionID       int64   `json:"emission_id"`
	EmissionQuantity int     `json:"emission_quantity"`
	EmissionAmount   float64 `json:"emission_amount"`
}

// NewEmissionSendResponse creates a new instance of EmissionSendResponse.
func NewEmissionSendResponse(httpCode int, status, message string, emissionID int64, emissionQuantity int, emissionAmount float64) EmissionSendResponse {
	return EmissionSendResponse{
		ResponseBase:     NewResponseBase(httpCode, status, message),
		EmissionID:       emissionID,
		EmissionQuantity: emissionQuantity,
		EmissionAmount:   emissionAmount,
	}
}

// Validate checks if the EmissionSendRequest has all required fields and valid data.
func (r *EmissionSendRequest) Validate(repo port.Repository) error {
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
func (r *EmissionSendRequest) valdateVendor(repo port.Repository) error {
	// Implement logic to validate if the vendor exists in the repository.
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
	r.vendorID = vendor.ID
	return nil
}

// validateDates checks if the provided dates are valid and in the correct format.
func (r *EmissionSendRequest) validateDates() error {
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
				errs = append(errs, fmt.Errorf("invoice_end_date cannot be before invoice_start_date"))
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
				errs = append(errs, fmt.Errorf("emission_date cannot be before invoice_end_date"))
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateDuplication checks if an emission for the given vendor and period already exists in the repository.
func (r *EmissionSendRequest) validateDuplication(repo port.Repository) error {
	// Implement logic to check for existing emissions in the repository.
	// This is a placeholder implementation; replace with actual database query.
	startDate, _ := time.Parse("2006-01-02", r.InvoiceStartDate)
	endDate, _ := time.Parse("2006-01-02", r.InvoiceEndDate)
	existingEmissions, err := repo.GetEmissions(r.vendorID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to check for existing emissions: %v", err)
	}
	if len(existingEmissions) > 0 {
		return fmt.Errorf("an emission for vendor '%s' and the specified period already exists", r.Vendor)
	}
	return nil
}

