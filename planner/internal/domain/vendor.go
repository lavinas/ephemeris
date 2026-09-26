package domain

import (
	"time"
)

// Vendor represents a supplier or service provider with their banking details and contact info.
type Vendor struct {
	ID            int64     `gorm:"primaryKey"`
	Nickname      string    `gorm:"not null;unique"`
	LegalName     string    `gorm:"not null"`
	TradingName   string    `gorm:"not null"`
	Document      string    `gorm:"not null;unique"`
	TaxDocument   string    `gorm:"not null"`
	AccountBank   string    `gorm:"not null"`
	AccountAgency string    `gorm:"not null"`
	AccountNumber string    `gorm:"not null"`
	PixToken      string    `gorm:"not null"`
	PixName       string    `gorm:"not null"`
	PixCity       string    `gorm:"not null"`
	LogoName      string    `gorm:"not null"`
	Email         string    `gorm:"not null"`
	Whatsapp      string    `gorm:"not null"`
	LastRps       int64     `gorm:"not null;default:0"`
	SmtpHost      string    `gorm:"not null"`
	SmtpPort      int       `gorm:"not null"`
	SmtpUser      string    `gorm:"not null"`
	SmtpPassword  string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null;default:now()"`
	UpdatedAt     time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for Vendor model.
func (Vendor) TableName() string {
	return "vendor"
}
