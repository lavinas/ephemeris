package dto

import (
	"time"
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
	BankName   string
	BankAgency string
	BankAccount string
	ReceiverName string
	ReceiverDocument string
}

// BillerPix represents the Pix information in the biller request.
type BillerPix struct {
	PixKey string
	ReceiverName string
	PixCopyPaste *string
	PixQRCode *string
} 

// BillerReceive represents the bank information in the biller request.
type BillerReceive struct {
	BankAccount *BillerBankAccount
	Pix         *BillerPix
}

// BillerRequest defines the interface for generating PDF files.
type BillerRequest struct {
	InvoiceID   string
	InvoiceDate time.Time
	InvoiceDue  time.Time
	Vendor      BillerVendor
	Customer    BillerCustomer
	Items       []BillerItem
	Receive	    BillerReceive
}
