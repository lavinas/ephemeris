package dto

import (
	"time"

	"billing/internal/port"
)

// InvoiceAutoBillRequest represents input to execute the auto billing process.
type InvoiceAutoBillRequest struct {
	DaysInAdvance int        `json:"days_in_advance"`
	ReferenceDate *time.Time `json:"reference_date,omitempty"`
}

// Validate validates the InvoiceAutoBillRequest.
func (r *InvoiceAutoBillRequest) Validate(repo port.Repository) error {
	return nil
}

// Reset resets the request fields.
func (r *InvoiceAutoBillRequest) Reset() {
	r.DaysInAdvance = 0
	r.ReferenceDate = nil
}

// InvoiceAutoBillResponse represents the result of the auto billing process.
type InvoiceAutoBillResponse struct {
	ResponseBase
	TotalFound  int      `json:"total_found"`
	TotalSent   int      `json:"total_sent"`
	TotalErrors int      `json:"total_errors"`
	Errors      []string `json:"errors,omitempty"`
}

// NewInvoiceAutoBillResponse creates a new InvoiceAutoBillResponse instance.
func NewInvoiceAutoBillResponse(statusCode int, statusMessage, errorMessage string,
	totalFound, totalSent, totalErrors int, errors []string) *InvoiceAutoBillResponse {
	return &InvoiceAutoBillResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		TotalFound:  totalFound,
		TotalSent:   totalSent,
		TotalErrors: totalErrors,
		Errors:      errors,
	}
}
