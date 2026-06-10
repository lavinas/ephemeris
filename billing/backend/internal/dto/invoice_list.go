package dto

import (
	"errors"
	"fmt"
	"time"

	"billing/internal/port"
)

// InvoiceListRequest represents the data transfer object for listing invoices with pagination and optional filters.
type InvoiceListRequest struct {
	Page             int     `json:"page" validate:"required,gt=0"`
	PageSize         int     `json:"page_size" validate:"required,gt=0"`
	Customer         *string `json:"customer,omitempty"`
	CustomerID       int64   `json:"customer_id,omitempty"`
	Vendor           *string `json:"vendor,omitempty"`
	VendorID         int64   `json:"vendor_id,omitempty"`
	InvoiceDate      *string `json:"invoicing,omitempty"`
	DueDate          *string `json:"due,omitempty"`
	EmailSentDate    *string `json:"email_sent,omitempty"`
	WhatsappSentDate *string `json:"whatsapp_sent,omitempty"`
	TaxDate          *string `json:"tax,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

// InvoiceListResponse represents the data transfer object for the response of listing invoices.
type InvoiceListResponse struct {
	ResponseBase
	Invoices []InvoiceList `json:"invoices,omitempty"`
}

// InvoiceList represents a single invoice item in the list response.
type InvoiceList struct {
	ID               int64                 `json:"id"`
	Customer         string                `json:"customer"`
	Amount           float64               `json:"amount"`
	InvoiceDate      string                `json:"invoicing"`
	DueDate          string                `json:"due"`
	EmailSentDate    string                `json:"email_sent"`
	WhatsappSentDate string                `json:"whatsapp_sent"`
	TaxDate          string                `json:"tax"`
	Notes            string                `json:"notes"`
	Items            []InvoiceListListItem `json:"items,omitempty"`
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

// NewInvoiceList creates a new instance of InvoiceList with the provided details.
func NewInvoiceList(id int64, customer string, amount float64, invoiceDate, dueDate time.Time,
	emailSentDate, whatsappSentDate, taxDate *time.Time, notes *string, items []InvoiceListListItem) InvoiceList {
	invoiceDateStr := invoiceDate.Format("2006-01-02")
	dueDateStr := dueDate.Format("2006-01-02")
	emailSentDateStr := "-"
	if emailSentDate != nil {
		emailSentDateStr = emailSentDate.Format("2006-01-02")
	}
	whatsappSentDateStr := "-"
	if whatsappSentDate != nil {
		whatsappSentDateStr = whatsappSentDate.Format("2006-01-02")
	}
	taxDateStr := "-"
	if taxDate != nil {
		taxDateStr = taxDate.Format("2006-01-02")
	}
	notesStr := "-"
	if notes != nil {
		notesStr = *notes
	}
	return InvoiceList{
		ID:               id,
		Customer:         customer,
		Amount:           amount,
		InvoiceDate:      invoiceDateStr,
		DueDate:          dueDateStr,
		EmailSentDate:    emailSentDateStr,
		WhatsappSentDate: whatsappSentDateStr,
		TaxDate:          taxDateStr,
		Notes:            notesStr,
		Items:            items,
	}
}

// NewInvoiceListListItem creates a new instance of InvoiceListListItem with the provided details.
func NewInvoiceListListItem(id int64, description string, quantity int,
	price float64) InvoiceListListItem {
	return InvoiceListListItem{
		ID:          id,
		Description: description,
		Quantity:    quantity,
		Price:       price,
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
	r.EmailSentDate = nil
	r.WhatsappSentDate = nil
	r.TaxDate = nil
	r.Notes = nil
}
