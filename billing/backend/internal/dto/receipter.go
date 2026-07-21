package dto

import (
	"html/template"
)

// ReceiptData represents the data required to generate a receipt.
type ReceiptData struct {
	VendorLogoBase64 template.URL
	VendorName       string
	VendorDocument   string
	VendorEmail      string
	VendorWhatsApp   string
	InvoiceNumber    string
	CustomerName     string
	CustomerDocument string
	CustomerEmail    string
	IssueDate        string
	DueDate          string
	PaymentDate      string
	Items            []ReceiptItem
	TotalAmount      float64
}

// ReceiptItem represents an individual item in the receipt.
type ReceiptItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Total       float64
}
