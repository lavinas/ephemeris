package domain

import (
	"time"
)

type InvoiceStatus struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

type InvoiceSendStatus struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

type InvoicePaymentStatus struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

type InvoiceItem struct {
	Base        `gorm:"embedded"`
	Invoice     *Invoice  `gorm:"foreignKey:ID, not null"`
	Contract    *Contract `gorm:"foreignKey:ID, not null"`
	Agenda      *Agenda   `gorm:"foreignKey:ID, not null"`
	Description string    `gorm:"type:varchar(100), not null"`
	Value       float64   `gorm:"type:numeric(20,2); not null"`
}

type Invoice struct {
	Base          `gorm:"embedded"`
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
