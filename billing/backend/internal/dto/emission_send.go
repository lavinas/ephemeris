package dto

import (
	"billing/internal/port"
)

// EmissionSendRequest represents the request data for sending emissions.
type EmissionSendRequest struct {
	VendorID         int64  `json:"vendor_id"`
	InvoiceStartDate string `json:"invoice_start_date"`
	InvoiceEndDate   string `json:"invoice_end_date"`
	EmissionDate     string `json:"emission_date"`
}

// EmissionSendResponse represents the response data after sending emissions.
type EmissionSendResponse struct {
	ResponseBase
	EmissionID int64 `json:"emission_id"`
	EmissionQuantity int   `json:"emission_quantity"`
	EmissionAmount   float64 `json:"emission_amount"`
}

// NewEmissionSendResponse creates a new instance of EmissionSendResponse.
func NewEmissionSendResponse(httpCode int, status, message string, emissionID int64, emissionQuantity int, emissionAmount float64) EmissionSendResponse {
	return EmissionSendResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		EmissionID:   emissionID,
		EmissionQuantity: emissionQuantity,
		EmissionAmount:   emissionAmount,
	}
}

// Validate checks if the EmissionSendRequest has all required fields and valid data.
func (r *EmissionSendRequest) Validate(repo port.Repository) error {
	// Implement validation logic for the request fields.
	return nil
}