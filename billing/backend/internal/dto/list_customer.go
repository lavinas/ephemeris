package dto

import (
	"errors"
	"fmt"
)

// ListCustomerRequest represents the data transfer object for listing customers with pagination.
type ListCustomerRequest struct {
	Page     int     `json:"page" validate:"required,gt=0"`
	PageSize int     `json:"page_size" validate:"required,gt=0"`
	Nickname *string `json:"nickname,omitempty"`
	Document *string `json:"document,omitempty"`
	Status   *int    `json:"status,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
}

// ListCustomerResponse represents the data transfer object for the response after listing customers.
type ListCustomerResponse struct {
	ResponseBase
	Customers []CustomerDTO `json:"customers"`
}

// CustomerDTO represents the data transfer object for a customer in the list response.
type CustomerDTO struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Nickname string  `json:"nickname"`
	Status   int     `json:"status"`
	Document *string `json:"document,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
}

// NewListCustomerResponse creates a new instance of ListCustomerResponse with the provided
// HTTP code, status, message, and customers.
func NewListCustomerResponse(httpCode int16, status, message string,
	customers []CustomerDTO) ListCustomerResponse {
	return ListCustomerResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		Customers:    customers,
	}
}

// Validate validates the ListCustomerRequest fields using the provided validator.
func (r *ListCustomerRequest) Validate() error {
	errs := make([]error, 0)
	if r.Page <= 0 {
		errs = append(errs, fmt.Errorf("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, fmt.Errorf("page_size must be greater than 0"))
	}
	return errors.Join(errs...)
}