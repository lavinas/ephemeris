package domain

import (
	"time"
)

// Invoice represents a billing invoice generated from a plan.
type Invoice struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement"`
	PlanID             int64          `gorm:"column:plan_id;not null;index"`
	Plan               *Plan          `gorm:"foreignKey:PlanID;references:ID"`
	InvoicingDate      time.Time      `gorm:"type:date;column:invoicing_date;not null"`
	DueDate            time.Time      `gorm:"type:date;column:due_date;not null"`
	PeriodStart        time.Time      `gorm:"type:date;column:period_start;not null"`
	PeriodEnd          time.Time      `gorm:"type:date;column:period_end;not null"`
	Amount             float64        `gorm:"type:numeric(15,2);column:amount;not null"`
	ReferenceInvoiceID *int64         `gorm:"column:reference_invoice_id"`
	CreatedAt          time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt          *time.Time     `gorm:"column:deleted_at;index"`
	Items              []InvoiceItem  `gorm:"foreignKey:InvoiceID;references:ID"`
}

// TableName specifies table name for Invoice.
func (Invoice) TableName() string {
	return "invoice"
}

// NewInvoice creates a new Invoice instance.
func NewInvoice(planID int64, invoicingDate, dueDate, periodStart, periodEnd time.Time, amount float64, refInvoiceID *int64) *Invoice {
	now := time.Now()
	return &Invoice{
		PlanID:             planID,
		InvoicingDate:      invoicingDate,
		DueDate:            dueDate,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		Amount:             amount,
		ReferenceInvoiceID: refInvoiceID,
		CreatedAt:          now,
		UpdatedAt:          now,
		DeletedAt:          nil,
	}
}

// InvoiceItem represents an individual line item within an invoice.
type InvoiceItem struct {
	ID               int64                `gorm:"primaryKey;autoIncrement"`
	InvoiceID        int64                `gorm:"column:invoice_id;not null;index"`
	Invoice          *Invoice             `gorm:"foreignKey:InvoiceID;references:ID"`
	ServiceID        int64                `gorm:"column:service_id;not null;index"`
	Service          *Service             `gorm:"foreignKey:ServiceID;references:ID"`
	Price            float64              `gorm:"type:numeric(15,2);column:price;not null"`
	Description      *string              `gorm:"type:varchar(255);column:description"`
	InvoicedQuantity int                  `gorm:"column:invoiced_quantity;not null"`
	RealizedQuantity int                  `gorm:"column:realized_quantity;not null"`
	CreatedAt        time.Time            `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time            `gorm:"column:updated_at;not null"`
	DeletedAt        *time.Time           `gorm:"column:deleted_at;index"`
	ItemSessions     []InvoiceItemSession `gorm:"foreignKey:InvoiceItemID;references:ID"`
}

// TableName specifies table name for InvoiceItem.
func (InvoiceItem) TableName() string {
	return "invoice_item"
}

// NewInvoiceItem creates a new InvoiceItem instance.
func NewInvoiceItem(invoiceID, serviceID int64, price float64, description *string, invoicedQty, realizedQty int) *InvoiceItem {
	now := time.Now()
	return &InvoiceItem{
		InvoiceID:        invoiceID,
		ServiceID:        serviceID,
		Price:            price,
		Description:      description,
		InvoicedQuantity: invoicedQty,
		RealizedQuantity: realizedQty,
		CreatedAt:        now,
		UpdatedAt:        now,
		DeletedAt:        nil,
	}
}

// InvoiceItemSession represents a link between an invoice item and a session.
type InvoiceItemSession struct {
	ID            int64        `gorm:"primaryKey;autoIncrement"`
	InvoiceItemID int64        `gorm:"column:invoice_item_id;not null;index"`
	InvoiceItem   *InvoiceItem `gorm:"foreignKey:InvoiceItemID;references:ID"`
	SessionID     int64        `gorm:"column:session_id;not null;index"`
	Session       *Session     `gorm:"foreignKey:SessionID;references:ID"`
	CreatedAt     time.Time    `gorm:"column:created_at;not null"`
}

// TableName specifies table name for InvoiceItemSession.
func (InvoiceItemSession) TableName() string {
	return "invoice_item_session"
}

// NewInvoiceItemSession creates a new InvoiceItemSession instance.
func NewInvoiceItemSession(invoiceItemID, sessionID int64) *InvoiceItemSession {
	return &InvoiceItemSession{
		InvoiceItemID: invoiceItemID,
		SessionID:     sessionID,
		CreatedAt:     time.Now(),
	}
}
