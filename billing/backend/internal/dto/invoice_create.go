package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/port"
)

const (
	notesLimit = 500
)

// InvoiceCreateRequest represents the data transfer object for creating a new invoice.
type InvoiceCreateRequest struct {
	Items []*InvoiceCreate `json:"items" validate:"required,dive"`
}

// InvoiceCreateRequest represents the data transfer object for creating a new invoice.
type InvoiceCreate struct {
	Vendor       string               `json:"vendor" validate:"required"`
	vendorID     int64                `json:"-"`
	Customer     string               `json:"customer" validate:"required"`
	customerID   int64                `json:"-"`
	InvoiceDate  string               `json:"invoicing" validate:"required"`
	DueDate      string               `json:"due" validate:"required"`
	InvoiceItems []*InvoiceCreateItem `json:"items" validate:"required,dive,required"`
	Note         string               `json:"note,omitempty"`
}

// InvoiceCreateItemDTO represents an item in the invoice with description, quantity, and unit price.
type InvoiceCreateItem struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

// InvoiceCreateResponse represents the data transfer object for the response
type InvoiceCreateResponse struct {
	ResponseBase
}

// NewInvoiceCreateResponse creates a new instance of InvoiceCreateResponse
func NewInvoiceCreateResponse(httpCode int16, status, message string) InvoiceCreateResponse {
	return InvoiceCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate validates the InvoiceCreateRequest fields using the provided validator.
func (r *InvoiceCreateRequest) Validate(repo port.Repository) error {
	if len(r.Items) == 0 {
		return errors.New("no invoice data provided")
	}
	errs := make([]error, 0)
	for i, item := range r.Items {
		if err := item.Validate(repo); err != nil {
			errs = append(errs, fmt.Errorf("invoice %d: %w", i, err))
		}
	}
	if len(errs) != 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// Validate validates the InvoiceCreateRequest fields using the provided validator.
func (r *InvoiceCreate) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateCustomer(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNotes(); err != nil {
		errs = append(errs, err)
	}
	if dateErrs := r.validateDates(); len(dateErrs) > 0 {
		errs = append(errs, dateErrs...)
	}
	if itemErrs := r.validateItems(); len(itemErrs) > 0 {
		errs = append(errs, itemErrs...)
	}
	if len(errs) != 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateNotes validates the note field to ensure it is not empty and does not exceed the limit.
func (r *InvoiceCreate) validateNotes() error {
	if r.Note == "" {
		return fmt.Errorf("note is required")
	}
	if len(r.Note) > notesLimit {
		return fmt.Errorf("note cannot exceed %d characters", notesLimit)
	}
	return nil
}

// validateDates validates the invoice and due dates to ensure they are in the correct format.
func (r *InvoiceCreate) validateDates() []error {
	errs := make([]error, 0)
	invoiceDate, err := time.Parse("2006-01-02", r.InvoiceDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("invoicing expected with YYYY-MM-DD format"))
	}
	dueDate, err := time.Parse("2006-01-02", r.DueDate)
	if err != nil {
		errs = append(errs, fmt.Errorf("due expected with YYYY-MM-DD format"))
	}
	if !invoiceDate.IsZero() && !dueDate.IsZero() && dueDate.Before(invoiceDate) {
		errs = append(errs, fmt.Errorf("due date cannot be before invoice date"))
	}
	return errs
}

// validateItems validates the invoice items to ensure they are not empty and have valid fields.
func (r *InvoiceCreate) validateItems() []error {
	errs := make([]error, 0)
	if len(r.InvoiceItems) == 0 {
		errs = append(errs, fmt.Errorf("at least one item is required"))
	}
	for i, item := range r.InvoiceItems {
		if err := item.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invoice item %d: %w", i, err))
		}
	}
	return errs
}

// validateCustomer validates the customer nickname and sets the customerID field if valid.
func (r *InvoiceCreate) validateCustomer(repo port.Repository) error {
	if r.Customer == "" {
		return fmt.Errorf("customer (nickname) is required")
	}
	customer, err := repo.GetCustomer(r.Customer)
	if err != nil {
		return fmt.Errorf("error finding customer: %v", err)
	}
	if customer == nil {
		return fmt.Errorf("customer not found: %s", r.Customer)
	}
	if r.vendorID != 0 && customer.VendorID != r.vendorID {
		return fmt.Errorf("customer '%s' does not belong to the specified vendor", r.Customer)
	}
	r.customerID = customer.ID
	return nil
}

// validateVendor validates the vendor nickname and sets the vendorID field if valid.
func (r *InvoiceCreate) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return fmt.Errorf("vendor (nickname) is required")
	}
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return fmt.Errorf("error finding vendor: %v", err)
	}
	if vendor == nil {
		return fmt.Errorf("vendor not found: %s", r.Vendor)
	}
	r.vendorID = vendor.ID
	return nil
}

// Validate validates the InvoiceCreateItem fields using the provided validator.
func (i *InvoiceCreateItem) Validate() error {
	errs := make([]error, 0)
	if i.Description == "" {
		errs = append(errs, fmt.Errorf("description is required"))
	}
	if i.Quantity <= 0 {
		errs = append(errs, fmt.Errorf("quantity must be greater than 0"))
	}
	if i.Price <= 0 {
		errs = append(errs, fmt.Errorf("price must be greater than 0"))
	}
	if len(errs) != 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain converts the InvoiceCreateRequest to a slice of domain.Invoice entities.
func (r *InvoiceCreateRequest) GetDomain() ([]domain.Invoice, error) {
	invoices := make([]domain.Invoice, len(r.Items))
	for i, item := range r.Items {
		invoices[i] = *item.GetDomain()
	}
	return invoices, nil
}

// GetDomain converts the InvoiceCreate to a domain.Invoice entity.
func (r *InvoiceCreate) GetDomain() *domain.Invoice {
	invoiceItems := make([]domain.InvoiceItem, len(r.InvoiceItems))
	for i, item := range r.InvoiceItems {
		invoiceItems[i] = *item.GetDomain()
	}
	iDate, _ := time.Parse("2006-01-02", r.InvoiceDate)
	dDate, _ := time.Parse("2006-01-02", r.DueDate)
	return domain.NewInvoice(r.customerID, iDate, dDate, r.Note, invoiceItems)
}

// GetDomain converts the InvoiceCreateItem to a domain.InvoiceItem entity.
func (i *InvoiceCreateItem) GetDomain() *domain.InvoiceItem {
	return domain.NewInvoiceItem(i.Description, i.Quantity, i.Price)
}
