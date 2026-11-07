package domain

import (
	"strings"
	"time"
)

// Invoice represents a billing invoice with customer details and amount.
type Invoice struct {
	ID               int64         `gorm:"primaryKey"`
	VendorID         int64         `gorm:"not null"`
	Vendor           Vendor        `gorm:"foreignKey:VendorID"`
	CustomerID       int64         `gorm:"not null"`
	Customer         Customer      `gorm:"foreignKey:CustomerID"`
	Amount           float64       `gorm:"not null"`
	InvoiceDate      time.Time     `gorm:"not null"`
	DueDate          time.Time     `gorm:"not null"`
	PaymentDate      *time.Time    `gorm:"null"`
	SentDate         *time.Time    `gorm:"null;column:sent_date"`
	ReceiptDate      *time.Time    `gorm:"null;column:receipt_date"`
	TaxDate          *time.Time    `gorm:"null"`
	CancellationDate *time.Time    `gorm:"null"`
	Notes            *string       `gorm:"null"`
	ExternalCode     *string       `gorm:"null;column:external_code"`
	CreatedAt        time.Time     `gorm:"autoCreateTime"`
	UpdatedAt        time.Time     `gorm:"autoUpdateTime"`
	InvoiceItems     []InvoiceItem `gorm:"foreignKey:InvoiceID"`
}

// NewInvoice creates a new Invoice instance with the details and calculates the total amount.
func NewInvoice(vendorID, customerID int64, invoiceDate, dueDate time.Time, paymentDate, cancellationDate *time.Time,
	notes, externalCode *string, items []InvoiceItem, amount ...*float64) *Invoice {
	invoice := &Invoice{
		VendorID:         vendorID,
		CustomerID:       customerID,
		InvoiceDate:      invoiceDate,
		DueDate:          dueDate,
		PaymentDate:      paymentDate,
		CancellationDate: cancellationDate,
		Notes:            notes,
		ExternalCode:     externalCode,
		InvoiceItems:     items,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if len(amount) > 0 && amount[0] != nil {
		invoice.Amount = *amount[0]
	} else {
		invoice.CalculateTotalAmount()
	}
	return invoice
}

// TableName specifies the table name for Invoice model.
func (Invoice) TableName() string {
	return "invoice"
}

// CalculateTotalAmount calculates the total amount of the invoice based on its items.
func (i *Invoice) CalculateTotalAmount() {
	var total float64
	for _, item := range i.InvoiceItems {
		total += float64(item.Quantity) * item.Price
	}
	i.Amount = total
}

// InvoiceItem represents an item in an invoice with description, quantity, and unit price.
type InvoiceItem struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	InvoiceID   int64     `json:"invoice_id" gorm:"not null"`
	Price       float64   `json:"price" gorm:"not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	NBSCode     *int64    `json:"nbs_code" gorm:"column:nbs_code"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewInvoiceItem creates a new InvoiceItem instance with the provided details.
func NewInvoiceItem(price float64, quantity int, description string, nbsCode ...*int64) *InvoiceItem {
	var nbs *int64
	if len(nbsCode) > 0 && nbsCode[0] != nil {
		nbs = nbsCode[0]
	}
	return &InvoiceItem{
		Price:       price,
		Quantity:    quantity,
		Description: description,
		NBSCode:     nbs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TableName specifies the table name for InvoiceItem model.
func (InvoiceItem) TableName() string {
	return "invoice_item"
}

// HasNBS checks if the invoice has items and all items have a valid NBS code.
func (i *Invoice) HasNBS() bool {
	if len(i.InvoiceItems) == 0 {
		return false
	}
	for _, item := range i.InvoiceItems {
		if item.NBSCode == nil || *item.NBSCode <= 0 {
			return false
		}
	}
	return true
}

// IsPF checks if the customer is an individual (Pessoa Física).
// Returns true if Document is present and has at most 11 digits (CPF).
func (i *Invoice) IsPF() bool {
	if i.Customer.Document == nil || *i.Customer.Document == "" {
		return false
	}
	cleanDoc := cleanNonDigits(*i.Customer.Document)
	return len(cleanDoc) > 0 && len(cleanDoc) <= 11
}

func cleanNonDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsOverdue checks if the invoice is overdue (dueDate <= today 00:00:00 and paymentDate is nil).
func (i *Invoice) IsOverdue() bool {
	if i.PaymentDate != nil {
		return false
	}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueStart := time.Date(i.DueDate.Year(), i.DueDate.Month(), i.DueDate.Day(), 23, 59, 59, 999, now.Location())
	return !dueStart.After(todayStart)
}

// IsTaxable checks if the invoice item is taxable and eligible for NFS-e emission.
// Requires payment confirmed, not cancelled, tax date nil, customer PF, and all items having NBS.
func (i *Invoice) IsTaxable() bool {
	if i.CancellationDate != nil {
		return false
	}
	if i.TaxDate != nil {
		return false
	}
	if i.PaymentDate == nil {
		return false
	}
	if !i.IsPF() {
		return false
	}
	if !i.HasNBS() {
		return false
	}
	return true
}

// CanDelete checks if the invoice can be deleted (all of CancellationDate, PaymentDate, TaxDate are nil).
func (i *Invoice) CanDelete() bool {
	return i.CancellationDate == nil &&
		i.PaymentDate == nil &&
		i.TaxDate == nil
}

// CanEdit checks if the invoice can be edited (CancellationDate is nil).
func (i *Invoice) CanEdit() bool {
	return i.CancellationDate == nil
}

// CanPay checks if the invoice can be paid (PaymentDate is nil).
func (i *Invoice) CanPay() bool {
	return i.PaymentDate == nil
}

// CanTax checks if the invoice can have a tax note emitted:
// must be paid, not cancelled, tax date nil, customer must be PF, and all items must have NBS.
func (i *Invoice) CanTax() bool {
	return i.CancellationDate == nil &&
		i.PaymentDate != nil &&
		i.TaxDate == nil &&
		i.IsPF() &&
		i.HasNBS()
}
