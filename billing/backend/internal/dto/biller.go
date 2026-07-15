package dto

import (
	"errors"
	"strings"
	"time"

	"billing/internal/port"
)

// BillerItem represents an item in the biller request.
type BillerItem struct {
	Description string
	Quantity    int
	Price       float64
}

// BillerVendor represents the vendor information in the biller request.
type BillerVendor struct {
	Logo     string
	Name     string
	Document string
	Address  *string
	Postcode *string
	City     *string
	State    *string
	Country  *string
	Email    *string
	Whatsapp *string
}

// BillerCustomer represents the customer information in the biller request.
type BillerCustomer struct {
	Name     string
	Document *string
	Email    *string
	Whatsapp *string
}

// BillerBankAccount represents the bank account information in the biller request.
type BillerBankAccount struct {
	BankName         string
	BankAgency       string
	BankAccount      string
	ReceiverName     string
	ReceiverDocument string
}

// BillerPix represents the Pix information in the biller request.
type BillerPix struct {
	PixKey       string
	ReceiverName string
	PixCopyPaste string
	PixQRCode    string
}

// BillerReceive represents the bank information in the biller request.
type BillerReceive struct {
	BankAccount *BillerBankAccount
	Pix         *BillerPix
}

// BillerRequest defines the interface for generating PDF files.
type BillerRequest struct {
	InvoiceID   int64
	InvoiceDate time.Time
	InvoiceDue  time.Time
	Vendor      BillerVendor
	Customer    BillerCustomer
	Items       []BillerItem
	Receive     BillerReceive
}

// Validate checks if the BillerRequest has valid data.
func (r *BillerRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)

	// Implement validation logic for BillerRequest fields if needed.
	if r.InvoiceID <= 0 {
		errs = append(errs, errors.New("invalid invoice ID: must be greater than 0"))
	}
	if r.InvoiceDate.IsZero() {
		errs = append(errs, errors.New("invalid invoice date: cannot be zero"))
	}
	if r.InvoiceDue.IsZero() {
		errs = append(errs, errors.New("invalid invoice due date: cannot be zero"))
	}
	if r.Vendor.Name == "" || r.Vendor.Document == "" {
		errs = append(errs, errors.New("invalid vendor information: name and document are required"))
	}
	if r.Customer.Name == "" {
		errs = append(errs, errors.New("invalid customer information: name is required"))
	}
	if len(r.Items) == 0 {
		errs = append(errs, errors.New("invalid items: at least one item is required"))
	}
	for _, item := range r.Items {
		if item.Description == "" || item.Quantity <= 0 || item.Price <= 0 {
			errs = append(errs, errors.New("invalid item: description, quantity, and price are required and must be greater than 0"))
		}
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// Reset resets the BillerRequest fields to their zero values.
func (r *BillerRequest) Reset() {
	r.InvoiceID = 0
	r.InvoiceDate = time.Time{}
	r.InvoiceDue = time.Time{}
	r.Vendor = BillerVendor{}
	r.Customer = BillerCustomer{}
	r.Items = nil
	r.Receive = BillerReceive{}
}
