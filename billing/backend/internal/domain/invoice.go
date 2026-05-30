package domain

import "time"

// Invoice represents a billing invoice with customer details and amount.
type Invoice struct {
	ID           int64         `gorm:"primaryKey"`
	BusinessID   int64         `gorm:"not null"`
	CustomerID   int64         `gorm:"not null"`
	Amount       float64       `gorm:"not null"`
	Notes        string        `gorm:"not null"`
	CreatedAt    time.Time     `gorm:"autoCreateTime"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime"`
	InvoiceItems []InvoiceItem `gorm:"foreignKey:InvoiceID"`
}

// NewInvoice creates a new Invoice instance with the provided details and calculates the total amount.
func NewInvoice(businessID, customerID int64, notes string, items []InvoiceItem) *Invoice {
	invoice := &Invoice{
		BusinessID:   businessID,
		CustomerID:   customerID,
		Notes:        notes,
		InvoiceItems: items,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	invoice.CalculateTotalAmount()
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
		total += float64(item.Quantity) * item.UnitPrice
	}
	i.Amount = total
}

// InvoiceItem represents an item in an invoice with description, quantity, and unit price.
type InvoiceItem struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	InvoiceID   int64     `json:"invoice_id" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	UnitPrice   float64   `json:"unit_price" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewInvoiceItem creates a new InvoiceItem instance with the provided details.
func NewInvoiceItem(description string, quantity int, unitPrice float64) *InvoiceItem {
	return &InvoiceItem{
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TableName specifies the table name for InvoiceItem model.
func (InvoiceItem) TableName() string {
	return "invoice_item"
}
