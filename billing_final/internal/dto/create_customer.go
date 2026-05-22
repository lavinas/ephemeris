package dto

import (
	"errors"
	"fmt"
)

// CreateCustomerRequest represents the request payload for creating a new customer.
type CreateCustomerRequest struct {
	Name     string `json:"name" validate:"required"`
	Nickname string `json:"nickname" validate:"required"`
	Document string `json:"document" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Whatsapp string `json:"whatsapp" validate:"required"`
}

// CreateCustomerResponse represents the response payload after creating a new customer.
type CreateCustomerResponse struct {
	ResponseBase
	ID int64 `json:"id"`
}

func (r *CreateCustomerRequest) Validate() error {
	errs := make([]error, 0)
	if r.Name == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if r.Nickname == "" {
		errs = append(errs, fmt.Errorf("nickname is required"))
	}
	return errors.Join(errs...)
}
