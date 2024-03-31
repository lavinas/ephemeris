package domain

import (
	"time"
)

type InvoiceStatus struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Name string `gorm:"type:varchar(100), not null"`
}

type InvoiceSendStatus struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Name string `gorm:"type:varchar(100), not null"`
}

type InvoicePaymentStatus struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Name string `gorm:"type:varchar(100), not null"`
}

type InvoiceItem struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Invoice     *Invoice  `gorm:"foreignKey:ID, not null"`
	Contract    *Contract `gorm:"foreignKey:ID, not null"`
	Agenda      *Agenda   `gorm:"foreignKey:ID, not null"`
	Description string    `gorm:"type:varchar(100), not null"`
	Value       float64   `gorm:"type:numeric(20,2); not null"`
}

type Invoice struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Ref           string                `gorm:"type:varchar(25); not null"`
	Client        *Client               `gorm:"foreignKey:ID, not null"`
	Date          time.Time             `gorm:"type:date; not null"`
	Due           time.Time             `gorm:"type:date; not null"`
	Value         float64               `gorm:"type:numeric(20,2); not null"`
	Status        *InvoiceStatus        `gorm:"foreignKey:ID, not null"`
	SendStatus    *InvoiceSendStatus    `gorm:"foreignKey:ID, not null"`
	SendDate      time.Time             `gorm:"type:date"`
	PaymentStatus *InvoicePaymentStatus `gorm:"foreignKey:ID, not null"`
	PaymentDate   time.Time             `gorm:"type:date"`
	PaymentValue  float64               `gorm:"type:numeric(20,2)"`
}
