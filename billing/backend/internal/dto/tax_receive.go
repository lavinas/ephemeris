package dto

import (
	"billing/internal/domain"
	"billing/internal/port"
	"errors"
	"fmt"
)

// TaxReceiveRequest represents the request data for emission receives.
type TaxReceiveRequest struct {
	Vendor     string           `json:"vendor"`
	VendorID   int64            `json:"-"`
	EmissionID int64            `json:"emission_id"`
	Source     string           `json:"source"`
	Emission   *domain.Emission `json:"-"`
}

// TaxReceiveResponse represents the response data after processing emission receives.
type TaxReceiveResponse struct {
	ResponseBase
	EmissionID       int64   `json:"emission_id"`
	EmissionQuantity int     `json:"emission_quantity"`
	EmissionAmount   float64 `json:"emission_amount"`
}

// NewTaxReceiveResponse creates a new instance of TaxReceiveResponse.
func NewTaxReceiveResponse(httpCode int, status, message string, emissionID int64,
	emissionQuantity int, emissionAmount float64) TaxReceiveResponse {
	return TaxReceiveResponse{
		ResponseBase:     NewResponseBase(httpCode, status, message),
		EmissionID:       emissionID,
		EmissionQuantity: emissionQuantity,
		EmissionAmount:   emissionAmount,
	}
}

// Validate checks if the TaxReceiveRequest has all required fields and valid data.
func (r *TaxReceiveRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmissionID(repo); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return err
	}
	return nil
}

// validateVendor checks if the vendor exists in the repository.
func (r *TaxReceiveRequest) validateVendor(repo port.Repository) error {
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

// validateEmissionID checks if the emission ID exists in the repository.
func (r *TaxReceiveRequest) validateEmissionID(repo port.Repository) error {
	if r.EmissionID <= 0 {
		return fmt.Errorf("emission_id must be a positive integer")
	}
	emission, err := repo.GetEmission(r.EmissionID)
	if err != nil {
		return fmt.Errorf("failed to validate emission_id: %v", err)
	}
	if emission == nil {
		return fmt.Errorf("emission_id '%d' does not exist", r.EmissionID)
	}
	if emission.VendorID != r.VendorID {
		return fmt.Errorf("emission_id '%d' does not belong to vendor '%s'",
			r.EmissionID, r.Vendor)
	}
	r.Emission = emission
	return nil
}

// GetDomain gets the domain representation of the TaxReceiveRequest.
func (r *TaxReceiveRequest) GetDomain() *domain.Emission {
	return r.Emission
}

// Reset resets the TaxReceiveRequest to its default state.
func (r *TaxReceiveRequest) Reset() {
	r.Vendor = ""
	r.VendorID = 0
	r.EmissionID = 0
	r.Source = ""
	r.Emission = nil
}
