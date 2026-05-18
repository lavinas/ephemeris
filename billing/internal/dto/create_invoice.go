package dto

import (
	"errors"
	"fmt"
)

// CreateInvoiceRequest represents the data transfer object for creating a new invoice.
type CreateInvoiceRequest struct {
	CustomerName     string                 `json:"customer_name" validate:"required"`
	CustomerEmail    string                 `json:"customer_email" validate:"required,email"`
	CustomerWhatsapp string                 `json:"customer_whatsapp" validate:"required"`
	CustomerDocument string                 `json:"customer_document" validate:"required"`
	Notes            string                 `json:"notes"`
	InvoiceItems     []CreateInvoiceItemDTO `json:"invoice_items" validate:"required,dive,required"`
}

// CreateInvoiceItemDTO represents an item in the invoice with description, quantity, and unit price.
type CreateInvoiceItemDTO struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// CreateInvoiceResponse represents the data transfer object for the response after creating a new invoice.
type CreateInvoiceResponse struct {
	ResponseBase
	ID int64 `json:"id"`
}

// Validate validates the CreateInvoiceRequest fields using the provided validator.
func (r *CreateInvoiceRequest) Validate() error {
	errs := make([]error, 0)
	if r.CustomerName == "" {
		errs = append(errs, fmt.Errorf("customer_name is required"))
	}
	if r.CustomerEmail == "" {
		errs = append(errs, fmt.Errorf("customer_email is required"))
	}
	if r.CustomerWhatsapp == "" {
		errs = append(errs, fmt.Errorf("customer_whatsapp is required"))
	}
	if r.CustomerDocument == "" {
		errs = append(errs, fmt.Errorf("customer_document is required"))
	}
	if len(r.InvoiceItems) == 0 {
		errs = append(errs, fmt.Errorf("at least one invoice item is required"))
	}
	for i, item := range r.InvoiceItems {
		if item.Description == "" {
			errs = append(errs, fmt.Errorf("description is required for invoice item %d", i))
		}
		if item.Quantity <= 0 {
			errs = append(errs, fmt.Errorf("quantity must be greater than 0 for invoice item %d", i))
		}
		if item.UnitPrice <= 0 {
			errs = append(errs, fmt.Errorf("unit_price must be greater than 0 for invoice item %d", i))
		}
	}
	return errors.Join(errs...)
}
