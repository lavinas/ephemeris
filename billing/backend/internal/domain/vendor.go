package domain

import (
	"time"
)

// Vendor represents a supplier or service provider with their banking details and contact info.
type Vendor struct {
	ID            int64     `gorm:"primaryKey"`
	LegalName     string    `gorm:"not null"`
	Nickname      string    `gorm:"not null;unique"`
	Document      string    `gorm:"not null;unique"`
	TaxDocument   string    `gorm:"not null"`
	AccountBank   string    `gorm:"not null"`
	AccountAgency string    `gorm:"not null"`
	AccountNumber string    `gorm:"not null"`
	PixToken      string    `gorm:"not null"`
	PixName       string    `gorm:"not null"`
	PixCity       string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null;default:now()"`
	UpdatedAt     time.Time `gorm:"not null;default:now()"`
	LastRps       int64     `gorm:"not null;default:0"`
}

// TableName specifies the table name for Vendor model.
func (Vendor) TableName() string {
	return "vendor"
}
