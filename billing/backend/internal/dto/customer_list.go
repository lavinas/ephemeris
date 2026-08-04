package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"billing/internal/port"
)

// CustomerListRequest represents the data transfer object for listing customers with pagination.
type CustomerListRequest struct {
	Page     int     `json:"page" validate:"required,gt=0"`
	PageSize int     `json:"page_size" validate:"required,gt=0"`
	Vendor   string  `json:"vendor" validate:"required"`
	VendorID int64   `json:"-" validate:"-"`
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
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	Name      string `json:"name"`
	Document  string `json:"document"`
	Email     string `json:"email"`
	Whatsapp  string `json:"whatsapp"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// NewCustomerListResponse creates a new instance of CustomerListResponse with the provided
// HTTP code, status, message, and customers.
func NewCustomerListResponse(httpCode int, status, message string,
	customers []CustomerDTO) CustomerListResponse {
	return CustomerListResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		Customers:    customers,
	}
}

// NewCustomerDTO creates a new instance of CustomerDTO from the given customer details.
func NewCustomerDTO(id int64, name, nickname string, status int,
	document, email, whatsapp *string, createdAt, updatedAt time.Time) CustomerDTO {
	docStr := "-"
	if document != nil {
		docStr = *document
	}
	emailStr := "-"
	if email != nil {
		emailStr = *email
	}
	whatsappStr := "-"
	if whatsapp != nil {
		whatsappStr = *whatsapp
	}

	return CustomerDTO{
		ID:        id,
		Nickname:  nickname,
		Name:      name,
		Document:  docStr,
		Email:     emailStr,
		Whatsapp:  whatsappStr,
		Status:    status,
		CreatedAt: createdAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: updatedAt.Format("2006-01-02 15:04:05"),
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
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if r.Status != nil && *r.Status != 1 && *r.Status != 0 && *r.Status != -1 {
		errs = append(errs, fmt.Errorf("status must be 1 (active), 0 (inactive), or -1 (all)"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateVendor checks if the provided vendor is valid and exists in the system.
func (r *CustomerListRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return errors.New("vendor is required")
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

// Reset resets the fields of the CustomerListRequest to their zero values.
func (r *CustomerListRequest) Reset() {
	r.Page = 0
	r.PageSize = 0
	r.Vendor = ""
	r.VendorID = 0
	r.Name = nil
	r.Nickname = nil
	r.Document = nil
	r.Status = nil
	r.Email = nil
	r.Whatsapp = nil
}
