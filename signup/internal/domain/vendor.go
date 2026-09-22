package domain

import (
	"fmt"
	"time"

	"signup/internal/port"
)

// Vendor represents a supplier or service provider with their banking details and contact info.
type Vendor struct {
	DomainBase
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

// NewVendor represents the input data required to create a new vendor.
func NewVendor(repo port.Repository, nickname, legalName, tradingName,
	document, taxDocument, accountBank, accountAgency, accountNumber,
	pixToken, pixName, pixCity, logoName, email, whatsapp,
	smtpHost string, smtpPort int, smtpUser, smtpPassword string) *Vendor {
	return &Vendor{
		DomainBase:    DomainBase{Repo: repo},
		Nickname:      nickname,
		LegalName:     legalName,
		TradingName:   tradingName,
		Document:      document,
		TaxDocument:   taxDocument,
		AccountBank:   accountBank,
		AccountAgency: accountAgency,
		AccountNumber: accountNumber,
		PixToken:      pixToken,
		PixName:       pixName,
		PixCity:       pixCity,
		LogoName:      logoName,
		Email:         email,
		Whatsapp:      whatsapp,
		SmtpHost:      smtpHost,
		SmtpPort:      smtpPort,
		SmtpUser:      smtpUser,
		SmtpPassword:  smtpPassword,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// StartVendor represents the input data required to create a new vendor.
func StartVendor(repo port.Repository) *Vendor {
	return &Vendor{DomainBase: DomainBase{Repo: repo}}
}

// TableName specifies the table name for Vendor model.
func (Vendor) TableName() string {
	return "vendor"
}

// Get methods for Vendor can be added here as needed.
func (v *Vendor) GetByNickname(nickname string) (bool, error) {
	conditions := map[string]interface{}{}
	conditions["nickname = ?"] = nickname
	resp, err := v.Repo.Find(v, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(resp) == 0 {
		return false, nil
	}
	vendor, ok := resp[0].(*Vendor)
	if !ok {
		return false, fmt.Errorf("failed to cast to Vendor")
	}
	*v = *vendor
	return true, nil
}

// GetByID retrieves a vendor by their ID.
func (v *Vendor) GetByID(id int64) (bool, error) {
	conditions := map[string]interface{}{}
	conditions["id = ?"] = id
	resp, err := v.Repo.Find(v, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(resp) == 0 {
		return false, nil
	}
	vendor, ok := resp[0].(*Vendor)
	if !ok {
		return false, fmt.Errorf("failed to cast to Vendor")
	}
	*v = *vendor
	return true, nil
}
