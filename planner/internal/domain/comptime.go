package domain

import (
	"time"
)

// NormalizeMonth returns the first day of the month for the given date.
func NormalizeMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EffectiveSessionMinutes calculates billable/consumed minutes according to status.
// Sessions with status "cancelada_nao_cobrar" must count as 0 consumed minutes.
func EffectiveSessionMinutes(minutes int, status string) int {
	if status == string(StatusCancelNotCharge) {
		return 0
	}
	return minutes
}

// CalculateSessionsFromMinutes converts a minute count into a fractional number of sessions
// based on the service's session duration in minutes (defaults to 60 if <= 0).
func CalculateSessionsFromMinutes(minutes, serviceMinutes int) float64 {
	if serviceMinutes <= 0 {
		serviceMinutes = 60
	}
	return float64(minutes) / float64(serviceMinutes)
}

// Comptime represents the monthly consolidated hour bank (banco de horas)
// for a specific vendor, customer and service.
type Comptime struct {
	ID              int64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	VendorID        int64                 `gorm:"column:vendor_id;not null;index" json:"vendor_id"`
	Vendor          *Vendor               `gorm:"foreignKey:VendorID;references:ID" json:"vendor,omitempty"`
	CustomerID      int64                 `gorm:"column:customer_id;not null;index" json:"customer_id"`
	Customer        *Customer             `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	ServiceID       int64                 `gorm:"column:service_id;not null;index" json:"service_id"`
	Service         *Service              `gorm:"foreignKey:ServiceID;references:ID" json:"service,omitempty"`
	Month           time.Time             `gorm:"type:date;column:month;not null;index" json:"month"`
	InvoicedMinutes int                   `gorm:"column:invoiced_minutes;not null;default:0" json:"invoiced_minutes"`
	ConsumedMinutes int                   `gorm:"column:consumed_minutes;not null;default:0" json:"consumed_minutes"`
	Balance         int                   `gorm:"column:balance;not null;default:0" json:"balance"`
	CreatedAt       time.Time             `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt       time.Time             `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt       *time.Time            `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	Sessions        []ComptimeSession     `gorm:"foreignKey:ComptimeID;references:ID" json:"sessions,omitempty"`
	InvoiceItems    []ComptimeInvoiceItem `gorm:"foreignKey:ComptimeID;references:ID" json:"invoice_items,omitempty"`
	Extras          []ComptimeExtra       `gorm:"foreignKey:ComptimeID;references:ID" json:"extras,omitempty"`
}

// TableName specifies table name for Comptime.
func (Comptime) TableName() string {
	return "comptime"
}

// NewComptime creates a new Comptime instance with zeroed minutes and balance.
func NewComptime(vendorID, customerID, serviceID int64, month time.Time) *Comptime {
	now := time.Now()
	return &Comptime{
		VendorID:        vendorID,
		CustomerID:      customerID,
		ServiceID:       serviceID,
		Month:           NormalizeMonth(month),
		InvoicedMinutes: 0,
		ConsumedMinutes: 0,
		Balance:         0,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
	}
}

// RecalculateBalance updates Balance given a previous month's balance.
func (c *Comptime) RecalculateBalance(previousBalance int) {
	c.Balance = previousBalance + c.InvoicedMinutes - c.ConsumedMinutes
	c.UpdatedAt = time.Now()
}

// MonthBalance returns the net minutes balance for the month (invoiced - consumed).
func (c *Comptime) MonthBalance() int {
	return c.InvoicedMinutes - c.ConsumedMinutes
}

// ComptimeSession tracks an individual session's debit against comptime.
type ComptimeSession struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ComptimeID      int64      `gorm:"column:comptime_id;not null;index" json:"comptime_id"`
	Comptime        *Comptime  `gorm:"foreignKey:ComptimeID;references:ID" json:"comptime,omitempty"`
	SessionID       int64      `gorm:"column:session_id;not null;index" json:"session_id"`
	Session         *Session   `gorm:"foreignKey:SessionID;references:ID" json:"session,omitempty"`
	Month           time.Time  `gorm:"type:date;column:month;not null;index" json:"month"`
	ConsumedMinutes int        `gorm:"column:consumed_minutes;not null;default:0" json:"consumed_minutes"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName specifies table name for ComptimeSession.
func (ComptimeSession) TableName() string {
	return "comptime_session"
}

// NewComptimeSession creates a new ComptimeSession instance.
func NewComptimeSession(comptimeID, sessionID int64, month time.Time, consumedMinutes int) *ComptimeSession {
	now := time.Now()
	return &ComptimeSession{
		ComptimeID:      comptimeID,
		SessionID:       sessionID,
		Month:           NormalizeMonth(month),
		ConsumedMinutes: consumedMinutes,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
	}
}

// ComptimeInvoiceItem tracks an individual invoice item's credit against comptime.
type ComptimeInvoiceItem struct {
	ID              int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	ComptimeID      int64        `gorm:"column:comptime_id;not null;index" json:"comptime_id"`
	Comptime        *Comptime    `gorm:"foreignKey:ComptimeID;references:ID" json:"comptime,omitempty"`
	InvoiceItemID   int64        `gorm:"column:invoice_item_id;not null;index" json:"invoice_item_id"`
	InvoiceItem     *InvoiceItem `gorm:"foreignKey:InvoiceItemID;references:ID" json:"invoice_item,omitempty"`
	Month           time.Time    `gorm:"type:date;column:month;not null;index" json:"month"`
	InvoicedMinutes int          `gorm:"column:invoiced_minutes;not null;default:0" json:"invoiced_minutes"`
	CreatedAt       time.Time    `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt       *time.Time   `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName specifies table name for ComptimeInvoiceItem.
func (ComptimeInvoiceItem) TableName() string {
	return "comptime_invoice_item"
}

// NewComptimeInvoiceItem creates a new ComptimeInvoiceItem instance.
func NewComptimeInvoiceItem(comptimeID, invoiceItemID int64, month time.Time, invoicedMinutes int) *ComptimeInvoiceItem {
	now := time.Now()
	return &ComptimeInvoiceItem{
		ComptimeID:      comptimeID,
		InvoiceItemID:   invoiceItemID,
		Month:           NormalizeMonth(month),
		InvoicedMinutes: invoicedMinutes,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
	}
}

// ComptimeExtra tracks manual minute adjustments (bonus, credit, discount).
type ComptimeExtra struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ComptimeID      int64      `gorm:"column:comptime_id;not null;index" json:"comptime_id"`
	Comptime        *Comptime  `gorm:"foreignKey:ComptimeID;references:ID" json:"comptime,omitempty"`
	Month           time.Time  `gorm:"type:date;column:month;not null;index" json:"month"`
	InvoicedMinutes int        `gorm:"column:invoiced_minutes;not null;default:0" json:"invoiced_minutes"`
	ConsumedMinutes int        `gorm:"column:consumed_minutes;not null;default:0" json:"consumed_minutes"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName specifies table name for ComptimeExtra.
func (ComptimeExtra) TableName() string {
	return "comptime_extra"
}

// NewComptimeExtra creates a new ComptimeExtra instance.
func NewComptimeExtra(comptimeID int64, month time.Time, invoicedMinutes, consumedMinutes int) *ComptimeExtra {
	now := time.Now()
	return &ComptimeExtra{
		ComptimeID:      comptimeID,
		Month:           NormalizeMonth(month),
		InvoicedMinutes: invoicedMinutes,
		ConsumedMinutes: consumedMinutes,
		CreatedAt:       now,
		UpdatedAt:       now,
		DeletedAt:       nil,
	}
}
