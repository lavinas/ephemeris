package domain

import "time"

// Invoice represents a billing invoice with customer details and amount.
type Invoice struct {
	ID               int64         `gorm:"primaryKey"`
	CustomerID       int64         `gorm:"not null"`
	Customer         Customer      `gorm:"foreignKey:CustomerID"`
	Amount           float64       `gorm:"not null"`
	InvoiceDate      time.Time     `gorm:"not null"`
	DueDate          time.Time     `gorm:"not null"`
	PaymentDate      *time.Time    `gorm:"null"`
	EmailSentDate    *time.Time    `gorm:"null"`
	WhatsappSentDate *time.Time    `gorm:"null"`
	TaxDate          *time.Time    `gorm:"null"`
	CancellationDate *time.Time    `gorm:"null"`
	Notes            *string       `gorm:"null"`
	CreatedAt        time.Time     `gorm:"autoCreateTime"`
	UpdatedAt        time.Time     `gorm:"autoUpdateTime"`
	InvoiceItems     []InvoiceItem `gorm:"foreignKey:InvoiceID"`
}

// NewInvoice creates a new Invoice instance with the details and calculates the total amount.
func NewInvoice(customerID int64, invoiceDate, dueDate time.Time, paymentDate, cancellationDate *time.Time,
	notes *string, items []InvoiceItem) *Invoice {
	invoice := &Invoice{
		CustomerID:       customerID,
		InvoiceDate:      invoiceDate,
		DueDate:          dueDate,
		PaymentDate:      paymentDate,
		CancellationDate: cancellationDate,
		Notes:            notes,
		InvoiceItems:     items,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewInvoiceItem creates a new InvoiceItem instance with the provided details.
func NewInvoiceItem(price float64, quantity int, description string) *InvoiceItem {
	return &InvoiceItem{
		Price:       price,
		Quantity:    quantity,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TableName specifies the table name for InvoiceItem model.
func (InvoiceItem) TableName() string {
	return "invoice_item"
}

// IsTaxable checks if the invoice item is taxable based on its description or other criteria.
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
	return true // Placeholder implementation; replace with actual logic.
}
