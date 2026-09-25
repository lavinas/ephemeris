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
	v.load(vendor)
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
	v.load(resp[0].(*Vendor))
	return true, nil
}
// GetByDocument retrieves a vendor by their document.
func (v *Vendor) GetByDocument(document string) (bool, error) {
	conditions := map[string]interface{}{}
	conditions["document = ?"] = document
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
	v.load(vendor)
	return true, nil
}

// Find retrieves vendors based on the specified conditions.
func (v *Vendor) Find(page, pageSize int) ([]port.Domain, error) {
	conditions := map[string]interface{}{}
	if v.Nickname != "" {
		conditions["nickname like ?"] = "%" + v.Nickname + "%"
	}
	if v.LegalName != "" {
		conditions["legal_name like ?"] = "%" + v.LegalName + "%"
	}
	if v.TradingName != "" {
		conditions["trading_name like ?"] = "%" + v.TradingName + "%"
	}
	if v.Document != "" {
		conditions["document like ?"] = "%" + v.Document + "%"
	}
	if v.Email != "" {
		conditions["email like ?"] = "%" + v.Email + "%"
	}
	if v.Whatsapp != "" {
		conditions["whatsapp like ?"] = "%" + v.Whatsapp + "%"
	}
	resp, err := v.Repo.Find(v, conditions, page, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]port.Domain, len(resp))
	for i, r := range resp {
		vendor, ok := r.(*Vendor)
		if !ok {
			return nil, fmt.Errorf("failed to cast to Vendor")
		}
		out[i] = vendor
	}
	return out, nil
}

// Validate validates vendor data before saving.
func (v *Vendor) Validate() error {
	if v.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	if err := v.ValidateNickname(v.Nickname); err != nil {
		return fmt.Errorf("nickname is invalid")
	}
	if v.LegalName == "" {
		return fmt.Errorf("legal_name is required")
	}
	if v.Document == "" {
		return fmt.Errorf("document is required")
	}
	if newDoc, err := v.ValidateCpfCnpj(v.Document); err != nil {
		return fmt.Errorf("document is invalid")
	} else {
		v.Document = newDoc
	}
	if v.Email != "" {
		if err := v.ValidateEmail(v.Email); err != nil {
			return fmt.Errorf("email is invalid")
		}
	}
	if v.Whatsapp != "" {
		if _, err := v.ValidateCellNumber(v.Whatsapp); err != nil {
			return fmt.Errorf("whatsapp is invalid")
		}
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	if v.UpdatedAt.IsZero() {
		v.UpdatedAt = time.Now()
	}
	return nil
}

// Save persists the vendor instance to the repository.
func (v *Vendor) Save() error {
	if err := v.Validate(); err != nil {
		return err
	}
	return v.Repo.Save(v)
}

// load loads all fields from a vendor
func (v *Vendor) load(vendor *Vendor) {
	v.ID = vendor.ID
	v.Nickname = vendor.Nickname
	v.LegalName = vendor.LegalName
	v.TradingName = vendor.TradingName
	v.Document = vendor.Document
	v.TaxDocument = vendor.TaxDocument
	v.AccountBank = vendor.AccountBank
	v.AccountAgency = vendor.AccountAgency
	v.AccountNumber = vendor.AccountNumber
	v.PixToken = vendor.PixToken
	v.PixName = vendor.PixName
	v.PixCity = vendor.PixCity
	v.LogoName = vendor.LogoName
	v.Email = vendor.Email
	v.Whatsapp = vendor.Whatsapp
	v.LastRps = vendor.LastRps
	v.SmtpHost = vendor.SmtpHost
	v.SmtpPort = vendor.SmtpPort
	v.SmtpUser = vendor.SmtpUser
	v.SmtpPassword = vendor.SmtpPassword
	v.CreatedAt = vendor.CreatedAt
	v.UpdatedAt = vendor.UpdatedAt
}
