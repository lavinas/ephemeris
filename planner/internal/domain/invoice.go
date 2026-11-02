package domain

import (
	"time"
)

// Invoice represents a billing invoice generated from a plan.
type Invoice struct {
	ID                 int64         `gorm:"primaryKey;autoIncrement"`
	PlanID             int64         `gorm:"column:plan_id;not null;index"`
	Plan               *Plan         `gorm:"foreignKey:PlanID;references:ID"`
	InvoicingDate      time.Time     `gorm:"type:date;column:invoicing_date;not null"`
	DueDate            time.Time     `gorm:"type:date;column:due_date;not null"`
	Amount             float64       `gorm:"type:numeric(15,2);column:amount;not null"`
	ReferenceInvoiceID *int64        `gorm:"column:reference_invoice_id"`
	CreatedAt          time.Time     `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time     `gorm:"column:updated_at;not null"`
	DeletedAt          *time.Time    `gorm:"column:deleted_at;index"`
	Items              []InvoiceItem `gorm:"foreignKey:InvoiceID;references:ID"`
}

// TableName specifies table name for Invoice.
func (Invoice) TableName() string {
	return "invoice"
}

// NewInvoice creates a new Invoice instance.
func NewInvoice(planID int64, invoicingDate, dueDate time.Time, amount float64, refInvoiceID *int64) *Invoice {
	now := time.Now()
	return &Invoice{
		PlanID:             planID,
		InvoicingDate:      invoicingDate,
		DueDate:            dueDate,
		Amount:             amount,
		ReferenceInvoiceID: refInvoiceID,
		CreatedAt:          now,
		UpdatedAt:          now,
		DeletedAt:          nil,
	}
}

// IsActiveInPeriod checks whether any active item of the invoice has a period [PeriodStart, PeriodEnd]
// that overlaps with [startDate, endDate].
func (inv *Invoice) IsActiveInPeriod(startDate, endDate time.Time) bool {
	if inv == nil || inv.DeletedAt != nil {
		return false
	}
	for idx := range inv.Items {
		if inv.Items[idx].IsActiveInPeriod(startDate, endDate) {
			return true
		}
	}
	return false
}

// InvoiceItem represents an individual line item within an invoice.
type InvoiceItem struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	InvoiceID   int64      `gorm:"column:invoice_id;not null;index"`
	Invoice     *Invoice   `gorm:"foreignKey:InvoiceID;references:ID"`
	ServiceID   int64      `gorm:"column:service_id;not null;index"`
	Service     *Service   `gorm:"foreignKey:ServiceID;references:ID"`
	Price       float64    `gorm:"type:numeric(15,2);column:price;not null"`
	Description *string    `gorm:"type:varchar(255);column:description"`
	Quantity    int        `gorm:"column:quantity;not null"`
	PeriodStart time.Time  `gorm:"type:date;column:period_start;not null"`
	PeriodEnd   time.Time  `gorm:"type:date;column:period_end;not null"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index"`
}

// TableName specifies table name for InvoiceItem.
func (InvoiceItem) TableName() string {
	return "invoice_item"
}

// NewInvoiceItem creates a new InvoiceItem instance.
func NewInvoiceItem(invoiceID, serviceID int64, price float64, description *string, quantity int, periodStart, periodEnd time.Time) *InvoiceItem {
	now := time.Now()
	return &InvoiceItem{
		InvoiceID:   invoiceID,
		ServiceID:   serviceID,
		Price:       price,
		Description: description,
		Quantity:    quantity,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}
}

// IsActiveInPeriod checks whether the item period [PeriodStart, PeriodEnd] overlaps with [startDate, endDate].
func (i *InvoiceItem) IsActiveInPeriod(startDate, endDate time.Time) bool {
	if i == nil || i.DeletedAt != nil || i.PeriodStart.IsZero() || i.PeriodEnd.IsZero() {
		return false
	}
	if startDate.IsZero() || endDate.IsZero() {
		return true
	}
	s := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	e := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
	pStart := time.Date(i.PeriodStart.Year(), i.PeriodStart.Month(), i.PeriodStart.Day(), 0, 0, 0, 0, time.UTC)
	pEnd := time.Date(i.PeriodEnd.Year(), i.PeriodEnd.Month(), i.PeriodEnd.Day(), 0, 0, 0, 0, time.UTC)
	return !pStart.After(e) && !pEnd.Before(s)
}
