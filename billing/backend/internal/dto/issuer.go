package dto

import (
	"fmt"
	"time"

	"html/template"

	"billing/internal/port"
)

// IssuerData represents the data required to generate a receipt.
type IssuerData struct {
	// vendor information
	VendorLogoBase64 template.URL
	VendorName       string
	VendorDocument   string
	VendorEmail      string
	VendorWhatsApp   string
	// vendor SMTP configuration
	VendorSMTPHost     string
	VendorSMTPPort     int
	VendorSMTPUsername string
	VendorSMTPPassword string
	// vendor Pix information
	VendorPixQRBase64  template.URL
	VendorPixCopyPaste string
	VendorPixName      string
	// vendor bank information
	VendorBank    string
	VendorAgency  string
	VendorAccount string
	// customer information
	CustomerNickname     string
	CustomerFirstName    string
	CustomerName         string
	CustomerDocumentType string
	CustomerDocument     string
	CustomerEmail        string
	// invoice information
	InvoiceNumber      int64
	InvoiceDate        time.Time
	InvoiceDueDate     time.Time
	InvoicePaymentDate time.Time
	InvoiceTotalAmount float64
	InvoiceItems       []ReceiptItem
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
	r.VendorLogoBase64 = template.URL("")
	r.VendorName = ""
	r.VendorDocument = ""
	r.VendorEmail = ""
	r.VendorWhatsApp = ""
	r.VendorSMTPHost = ""
	r.VendorSMTPPort = 0
	r.VendorSMTPUsername = ""
	r.VendorSMTPPassword = ""
	r.VendorPixQRBase64 = template.URL("")
	r.VendorPixCopyPaste = ""
	r.VendorPixName = ""
	r.VendorBank = ""
	r.VendorAgency = ""
	r.VendorAccount = ""
	r.CustomerNickname = ""
	r.CustomerFirstName = ""
	r.CustomerName = ""
	r.CustomerDocumentType = ""
	r.CustomerDocument = ""
	r.CustomerEmail = ""
	r.InvoiceNumber = 0
	r.InvoiceDate = time.Time{}
	r.InvoiceDueDate = time.Time{}
	r.InvoicePaymentDate = time.Time{}
	r.InvoiceTotalAmount = 0.0
}

// Validate checks if the IssuerData fields are valid and returns an error if any required field is missing or invalid.
func (r *IssuerData) Validate(repo port.Repository) error {
	errors := []error{}
	if r.VendorLogoBase64 == "" {
		errors = append(errors, fmt.Errorf("vendor logo is required"))
	}
	if r.VendorEmail == "" {
		errors = append(errors, fmt.Errorf("vendor email is required"))
	}
	if r.VendorSMTPHost == "" {
		errors = append(errors, fmt.Errorf("vendor SMTP host is required"))
	}
	if r.VendorSMTPPort <= 0 {
		errors = append(errors, fmt.Errorf("vendor SMTP port must be greater than zero"))
	}
	if r.VendorSMTPUsername == "" {
		errors = append(errors, fmt.Errorf("vendor SMTP username is required"))
	}
	if r.VendorSMTPPassword == "" {
		errors = append(errors, fmt.Errorf("vendor SMTP password is required"))
	}
	if r.VendorName == "" {
		errors = append(errors, fmt.Errorf("vendor name is required"))
	}
	if r.VendorDocument == "" {
		errors = append(errors, fmt.Errorf("vendor document is required"))
	}
	if r.CustomerNickname == "" {
		errors = append(errors, fmt.Errorf("customer nickname is required"))
	}
	if r.CustomerName == "" {
		errors = append(errors, fmt.Errorf("customer name is required"))
	}
	if r.InvoiceNumber <= 0 {
		errors = append(errors, fmt.Errorf("invoice number must be greater than zero"))
	}
	if r.InvoiceDate.IsZero() {
		errors = append(errors, fmt.Errorf("invoice date is required"))
	}
	if r.InvoiceDueDate.IsZero() {
		errors = append(errors, fmt.Errorf("due date is required"))
	}
	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %v", errors)
	}
	return nil
}
