package dto

import (
	"billing/internal/domain"
	"billing/internal/port"
	"encoding/base64"
	"errors"
	"fmt"
)

// TaxClearRequest represents the request data for emission clears.
type TaxClearRequest struct {
	Vendor       string           `json:"vendor"`
	VendorID     int64            `json:"-"`
	EmissionID   int64            `json:"emission_id"`
	Base64Source string           `json:"source"`
	Emission     *domain.Emission `json:"-"`
}

// TaxClearResponse represents the response data after processing emission clears.
type TaxClearResponse struct {
	ResponseBase
}

// NewTaxClearResponse creates a new instance of TaxClearResponse.
func NewTaxClearResponse(httpCode int, status, message string) TaxClearResponse {
	return TaxClearResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the TaxClearRequest has all required fields and valid data.
func (r *TaxClearRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmissionID(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSource(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return err
	}
	return nil
}

// validateVendor checks if the vendor exists in the repository.
func (r *TaxClearRequest) validateVendor(repo port.Repository) error {
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
func (r *TaxClearRequest) validateEmissionID(repo port.Repository) error {
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

// validateSource checks if the Base64Source is a valid base64 string.
func (r *TaxClearRequest) validateSource() error {
	if r.Base64Source == "" {
		return fmt.Errorf("source is required")
	}
	_, err := base64.StdEncoding.DecodeString(r.Base64Source)
	if err != nil {
		return fmt.Errorf("invalid base64 source: %v", err)
	}
	return nil
}

// GetDomain gets the domain representation of the TaxClearRequest.
func (r *TaxClearRequest) GetDomain() *domain.Emission {
	return r.Emission
}

// Reset resets the TaxClearRequest to its default state.
func (r *TaxClearRequest) Reset() {
	r.Vendor = ""
	r.VendorID = 0
	r.EmissionID = 0
	r.Base64Source = ""
	r.Emission = nil
}
