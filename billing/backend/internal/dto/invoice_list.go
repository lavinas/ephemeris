package dto

import (
	"errors"
	"fmt"

	"billing/internal/port"
)

// InvoiceListRequest represents the data transfer object for listing invoices with pagination and optional filters.
type InvoiceListRequest struct {
	Page        int     `json:"page" validate:"required,gt=0"`
	PageSize    int     `json:"page_size" validate:"required,gt=0"`
	Customer    *string `json:"customer,omitempty"`
	CustomerID  int64   `json:"customer_id,omitempty"`
	Vendor      *string `json:"vendor,omitempty"`
	VendorID    int64   `json:"vendor_id,omitempty"`
	InvoiceDate *string `json:"invoicing,omitempty"`
	DueDate     *string `json:"due,omitempty"`
}

// InvoiceListResponse represents the data transfer object for the response of listing invoices.
type InvoiceListResponse struct {
	ResponseBase
	Vendor   string        `json:"vendor"`
	Invoices []InvoiceList `json:"invoices,omitempty"`
}

// InvoiceList represents a single invoice item in the list response.
type InvoiceList struct {
	ID          int64                 `json:"id"`
	Customer    string                `json:"customer"`
	Amount      float64               `json:"amount"`
	InvoiceDate string                `json:"invoicing"`
	DueDate     string                `json:"due"`
	Status      string                `json:"status"`
	Notes       string                `json:"notes"`
	Items       []InvoiceListListItem `json:"items,omitempty"`
}

// InvoiceListListItem represents a single invoice item in the list response.
type InvoiceListListItem struct {
	ID          int64   `json:"id"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

// NewInvoiceListResponse creates a new instance of InvoiceListResponse
func NewInvoiceListResponse(code int, status, message string,
	items []InvoiceList) InvoiceListResponse {
	return InvoiceListResponse{
		ResponseBase: NewResponseBase(code, status, message),
		Invoices:     items,
	}
}

// Validate validates the InvoiceListRequest fields using the provided validator.
func (r *InvoiceListRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if r.Page <= 0 {
		errs = append(errs, errors.New("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, errors.New("page_size must be greater than 0"))
	}
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateCustomer(repo); err != nil {
		errs = append(errs, err)
	}
	// Validation logic can be implemented here if needed
	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %v", errs)
	}
	return nil
}

// validateVendor validates the vendor field to ensure it is not empty and exists in the repository.
func (r *InvoiceListRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == nil || *r.Vendor == "" {
		return fmt.Errorf("vendor cannot be empty")
	}
	vendors, err := repo.GetVendor(*r.Vendor)
	if err != nil {
		return fmt.Errorf("vendor not found: %v", err)
	}
	if vendors == nil {
		return fmt.Errorf("vendor not found")
	}
	r.VendorID = vendors.ID
	return nil
}

// validateCustomer validates the customer field to ensure it exists in the repository.
func (r *InvoiceListRequest) validateCustomer(repo port.Repository) error {
	if r.Customer == nil || *r.Customer == "" {
		r.CustomerID = 0 // Set to 0 to indicate no filter on customer
		return nil
	}
	customers, err := repo.GetCustomer(r.VendorID, *r.Customer)
	if err != nil {
		return fmt.Errorf("customer not found: %v", err)
	}
	if customers == nil {
		return fmt.Errorf("customer not found")
	}
	if r.VendorID != 0 && customers.VendorID != r.VendorID {
		return fmt.Errorf("customer '%s' does not belong to the specified vendor", *r.Customer)
	}
	r.CustomerID = customers.ID
	return nil
}

// Reset resets the fields of the InvoiceListRequest to their zero values.
func (r *InvoiceListRequest) Reset() {
	r.Page = 0
	r.PageSize = 0
	r.Customer = nil
	r.CustomerID = 0
	r.Vendor = nil
	r.VendorID = 0
	r.InvoiceDate = nil
	r.DueDate = nil
}
