package dto

import (
	"errors"
	"fmt"
	"strings"

	"billing/internal/port"
)

// CustomerListRequest represents the data transfer object for listing customers with pagination.
type CustomerListRequest struct {
	Page     int     `json:"page" validate:"required,gt=0"`
	PageSize int     `json:"page_size" validate:"required,gt=0"`
	Name     *string `json:"name,omitempty"`
	Nickname *string `json:"nickname,omitempty"`
	Document *string `json:"document,omitempty"`
	Status   *int    `json:"status,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
}

// CustomerListResponse represents the data transfer object for the response after listing customers.
type CustomerListResponse struct {
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

// NewCustomerListResponse creates a new instance of CustomerListResponse with the provided
// HTTP code, status, message, and customers.
func NewCustomerListResponse(httpCode int16, status, message string,
	customers []CustomerDTO) CustomerListResponse {
	return CustomerListResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		Customers:    customers,
	}
}

// NewCustomerDTO creates a new instance of CustomerDTO from the given customer details.
func NewCustomerDTO(id int64, name, nickname string, status int,
	document, email, whatsapp *string) CustomerDTO {
	return CustomerDTO{
		ID:       id,
		Name:     name,
		Nickname: nickname,
		Status:   status,
		Document: document,
		Email:    email,
		Whatsapp: whatsapp,
	}
}

// Validate validates the CustomerListRequest fields using the provided validator.
func (r *CustomerListRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if r.Page <= 0 {
		errs = append(errs, fmt.Errorf("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, fmt.Errorf("page_size must be greater than 0"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}
