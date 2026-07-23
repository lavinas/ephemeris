package dto

import (
	"fmt"

	"html/template"

	"billing/internal/port"
)

// IssuerData represents the data required to generate a receipt.
type IssuerData struct {
	VendorLogoBase64   template.URL
	VendorName         string
	VendorDocument     string
	VendorEmail        string
	VendorWhatsApp     string
	VendorSMTPHost     string
	VendorSMTPPort     int
	VendorSMTPUsername string
	VendorSMTPPassword string
	InvoiceNumber      string
	CustomerName       string
	CustomerDocument   string
	CustomerEmail      string
	IssueDate          string
	DueDate            string
	PaymentDate        string
	Items              []ReceiptItem
	TotalAmount        float64
}

// ReceiptItem represents an individual item in the receipt.
type ReceiptItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Total       float64
}

// Reset resets the IssuerData fields to their zero values.
func (r *IssuerData) Reset() {
	r.VendorLogoBase64 = ""
	r.VendorName = ""
	r.VendorDocument = ""
	r.VendorEmail = ""
	r.VendorWhatsApp = ""
	r.VendorSMTPHost = ""
	r.VendorSMTPPort = 0
	r.VendorSMTPUsername = ""
	r.VendorSMTPPassword = ""
	r.InvoiceNumber = ""
	r.CustomerName = ""
	r.CustomerDocument = ""
	r.CustomerEmail = ""
	r.IssueDate = ""
	r.DueDate = ""
	r.PaymentDate = ""
	r.Items = nil
	r.TotalAmount = 0.0
}

// Validate checks if the IssuerData fields are valid and returns an error if any required field is missing or invalid.
func (r *IssuerData) Validate(repo port.Repository) error {
	errors := []error{}
	if r.VendorName == "" {
		errors = append(errors, fmt.Errorf("vendor name is required"))
	}
	if r.VendorDocument == "" {
		errors = append(errors, fmt.Errorf("vendor document is required"))
	}
	if r.CustomerName == "" {
		errors = append(errors, fmt.Errorf("customer name is required"))
	}
	if r.CustomerDocument == "" {
		errors = append(errors, fmt.Errorf("customer document is required"))
	}
	if r.InvoiceNumber == "" {
		errors = append(errors, fmt.Errorf("invoice number is required"))
	}
	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %v", errors)
	}
	return nil
}
