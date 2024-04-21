package domain

// InvoiceItem represents the invoice item entity
type InvoiceItem struct {
	ID          string  `gorm:"type:varchar(25); primaryKey"`
	InvoiceID   string  `gorm:"type:varchar(25); not null"`
	AgendaID    string  `gorm:"type:varchar(25); not null"`
	Value       float64 `gorm:"type:numeric(20,2); not null"`
	Description string  `gorm:"type:varchar(100); not null"`
}
