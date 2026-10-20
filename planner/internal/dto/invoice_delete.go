package dto

import (
	"encoding/json"
	"errors"
	"planner/internal/port"
)

// InvoiceDeleteRequest represents a request to soft-delete an invoice.
type InvoiceDeleteRequest struct {
	InvoiceID int64 `json:"id" validate:"required"`
}

// UnmarshalJSON supports custom unmarshaling for both "id" and "invoice_id".
func (r *InvoiceDeleteRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	getInt64 := func(keys ...string) int64 {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					return int64(val)
				case int64:
					return val
				case int:
					return int64(val)
				}
			}
		}
		return 0
	}

	r.InvoiceID = getInt64("id", "invoice_id", "fatura_id")
	return nil
}

// InvoiceDeleteResponse represents the response after deleting an invoice.
type InvoiceDeleteResponse struct {
	ResponseBase
}

// NewInvoiceDeleteResponse creates a new instance of InvoiceDeleteResponse.
func NewInvoiceDeleteResponse(statusCode int, statusMessage string, errorMessage string) *InvoiceDeleteResponse {
	return &InvoiceDeleteResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
	}
}

// Validate checks if the InvoiceDeleteRequest has valid data.
func (r *InvoiceDeleteRequest) Validate(repo port.Repository) error {
	if r.InvoiceID <= 0 {
		return errors.New("id da fatura inválido: deve ser maior que zero")
	}
	return nil
}
