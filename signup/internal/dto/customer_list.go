package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"signup/internal/domain"
	"signup/internal/port"
)

// CustomerListRequest represents the data transfer object for listing customers with pagination.
type CustomerListRequest struct {
	RequestBase `json:"-" validate:"-"`
	Page        int     `json:"page" validate:"required,gt=0"`
	PageSize    int     `json:"page_size" validate:"required,gt=0"`
	Vendor      string  `json:"vendor" validate:"required"`
	VendorID    int64   `json:"-" validate:"-"`
	Name        *string `json:"name,omitempty"`
	Nickname    *string `json:"nickname,omitempty"`
	Document    *string `json:"document,omitempty"`
	Status      *int    `json:"status,omitempty"`
	Email       *string `json:"email,omitempty"`
	Whatsapp    *string `json:"whatsapp,omitempty"`
}

// NewCustomerListRequest creates a new instance of CustomerListRequest with the provided repository.
func NewCustomerListRequest(repo port.Repository) *CustomerListRequest {
	return &CustomerListRequest{
		RequestBase: NewRequestBase(repo),
	}
}

// CustomerListResponse represents the data transfer object for the response after listing customers.
type CustomerListResponse struct {
	ResponseBase
	Customers []CustomerDTO `json:"customers"`
}

// CustomerDTO represents the data transfer object for a customer in the list response.
type CustomerDTO struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Document string `json:"document"`
	Email    string `json:"email"`
	Whatsapp string `json:"whatsapp"`
	Status   int    `json:"status"`
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
		ID:       id,
		Nickname: nickname,
		Name:     name,
		Document: docStr,
		Email:    emailStr,
		Whatsapp: whatsappStr,
		Status:   status,
	}
}

// Validate validates the CustomerListRequest fields using the provided validator.
func (r *CustomerListRequest) Validate() error {
	errs := make([]error, 0)
	if r.Page <= 0 {
		errs = append(errs, fmt.Errorf("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, fmt.Errorf("page_size must be greater than 0"))
	}
	if err := r.validateVendor(r.Repo); err != nil {
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

// GetDomain converts the CustomerListRequest to a domain.Customer entity.
func (r *CustomerListRequest) GetDomain() (port.Domain, error) {
	customer := domain.StartCustomer(r.Repo)
	customer.VendorID = r.VendorID
	if r.Name != nil {
		customer.Name = *r.Name
	}
	if r.Nickname != nil {
		customer.Nickname = *r.Nickname
	}
	customer.Document = r.Document
	customer.Email = r.Email
	customer.Whatsapp = r.Whatsapp
	customer.Status = r.Status
	return customer, nil
}

// GetOutDTO constructs an output DTO for the customer list request.
func (r *CustomerListRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	var customers []CustomerDTO
	if list, ok := data.([]port.Domain); ok {
		for _, d := range list {
			if c, ok := d.(*domain.Customer); ok {
				customers = append(customers, NewCustomerDTO(c.ID, c.Name, c.Nickname, *c.Status, c.Document, c.Email, c.Whatsapp, c.CreatedAt, c.UpdatedAt))
			}
		}
		return NewCustomerListResponse(httpCode, status, message, customers)
	}
	return NewCustomerListResponse(httpCode, status, message, []CustomerDTO{})
}

// GetPageParams returns the pagination parameters.
// returns (page, pageSize)
func (r *CustomerListRequest) GetPageParams() (int, int) {
	return r.Page, r.PageSize
}

// validateVendor checks if the provided vendor is valid and exists in the system.
func (r *CustomerListRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return errors.New("vendor is required")
	}
	vendor := domain.StartVendor(repo)
	if ok, err := vendor.GetByNickname(r.Vendor); err != nil {
		return fmt.Errorf("failed to validate vendor: %v", err)
	} else if !ok {
		return fmt.Errorf("vendor '%s' does not exist", r.Vendor)
	}
	r.VendorID = vendor.ID
	return nil
}
