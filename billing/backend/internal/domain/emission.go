package domain

import (
	"time"
)

// Emission represents an emission record associated with a vendor, including the period of the emission.
type Emission struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	VendorID      int64           `gorm:"not null;index"`
	Vendor        Vendor          `gorm:"foreignKey:VendorID"`
	EmissionDate  time.Time       `gorm:"not null;date"`
	PeriodStart   time.Time       `gorm:"date"`
	PeriodEnd     time.Time       `gorm:"date"`
	RPSStart      int64           `gorm:"index"`
	RPSEnd        int64           `gorm:"index"`
	NFEStart      *int64          `gorm:"index"`
	NFEEnd        *int64          `gorm:"index"`
	NFEDatetime   *time.Time      `gorm:"date"`
	Amount        float64         `gorm:"not null"`
	Quantity      int             `gorm:"not null"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
	EmissionItems []*EmissionItem `gorm:"foreignKey:EmissionID"`
}

// EmissionItem represents an individual emission item with a date and amount.
type EmissionItem struct {
	ID              int64      `gorm:"primaryKey;autoIncrement"`
	EmissionID      int64      `gorm:"not null;index"`
	InvoiceID       int64      `gorm:"not null;index"`
	Invoice         Invoice    `gorm:"foreignKey:InvoiceID"`
	RPSNumber       int64      `gorm:"not null;index"`
	NFENumber       *int64     `gorm:"index"`
	NFEDatetime     *time.Time `gorm:"date"`
	NFEVerification *string    `gorm:"size:255"`
}

// NewEmission creates a new Emission instance with the provided details.
func NewEmission(vendor *Vendor, periodStart, periodEnd, emissionDate time.Time, RPSStart,
	RPSEnd int64, amount float64, quantity int, items []*EmissionItem) *Emission {
	return &Emission{
		VendorID:      vendor.ID,
		Vendor:        *vendor,
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		EmissionDate:  emissionDate,
		RPSStart:      RPSStart,
		RPSEnd:        RPSEnd,
		NFEStart:      nil,
		NFEEnd:        nil,
		Amount:        amount,
		Quantity:      quantity,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		EmissionItems: items,
	}
}

// TableName specifies the table name for Emission model.
func (Emission) TableName() string {
	return "emission"
}

// NewEmissionItem creates a new EmissionItem instance with the provided details.
func NewEmissionItem(emissionID, invoiceID, RPSNum int64, invoice Invoice) *EmissionItem {
	return &EmissionItem{
		EmissionID: emissionID,
		InvoiceID:  invoiceID,
		Invoice:    invoice,
		RPSNumber:  RPSNum,
	}
}

// TableName specifies the table name for EmissionItem model.
func (EmissionItem) TableName() string {
	return "emission_item"
}
