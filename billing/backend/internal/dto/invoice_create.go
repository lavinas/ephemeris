package dto

import (
	"errors"
	"fmt"
	"strings"

	"billing/internal/domain"
	"billing/internal/port"
)


// InvoiceCreateRequest represents the data transfer object for creating a new invoice.
type InvoiceCreateRequest struct {
	Items []InvoiceCreate `json:"items" validate:"required,dive"`
}


// InvoiceCreateRequest represents the data transfer object for creating a new invoice.
type InvoiceCreate struct {
	Business         string              `json:"business" validate:"required"`
	Customer         string              `json:"customer_name" validate:"required"`
	InvoiceItems     []InvoiceCreateItem `json:"invoice_items" validate:"required,dive,required"`
	Note             string              `json:"note,omitempty"`
}

// InvoiceCreateItemDTO represents an item in the invoice with description, quantity, and unit price.
type InvoiceCreateItem struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
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
	if r.Customer == "" {
		errs = append(errs, fmt.Errorf("customer_name is required"))
	}
	if r.Business == "" {
		errs = append(errs, fmt.Errorf("business is required"))
	}
	if len(r.InvoiceItems) == 0 {
		errs = append(errs, fmt.Errorf("at least one invoice item is required"))
	}
	for i, item := range r.InvoiceItems {
		if err := item.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invoice item %d: %w", i, err))
		}
	}
	if len(errs) != 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
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
	if i.UnitPrice <= 0 {
		errs = append(errs, fmt.Errorf("unit_price must be greater than 0"))
	}
	if len(errs) != 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain converts the InvoiceCreate to a domain.Invoice entity.
func (r *InvoiceCreate) GetDomain(businessID, customerID int64) *domain.Invoice {
	invoiceItems := make([]domain.InvoiceItem, len(r.InvoiceItems))
	for i, item := range r.InvoiceItems {
		invoiceItems[i] = *item.GetDomain()
	}
	return domain.NewInvoice(businessID, customerID, r.Note, invoiceItems)
}

// GetDomain converts the InvoiceCreateItem to a domain.InvoiceItem entity.
func (i *InvoiceCreateItem) GetDomain() *domain.InvoiceItem {
	return domain.NewInvoiceItem(i.Description, i.Quantity, i.UnitPrice)
}




