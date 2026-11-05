package domain

import "time"

// Vendor represents an organization/company using the system
type Vendor struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Nickname      string     `gorm:"column:nickname;size:150;unique" json:"nickname"`
	LegalName     string     `gorm:"column:legal_name;size:150;not null" json:"legal_name"`
	TradingName   string     `gorm:"column:trading_name;size:150" json:"trading_name"`
	Document      string     `gorm:"column:document;size:50;not null;unique" json:"document"`
	Email         string     `gorm:"column:email;size:150;not null" json:"email"`
	Whatsapp      string     `gorm:"column:whatsapp;size:20;not null" json:"whatsapp"`
	TaxDocument   string     `gorm:"column:tax_document;size:50;not null" json:"tax_document"`
	AccountBank   string     `gorm:"column:account_bank;size:100;not null" json:"account_bank"`
	AccountAgency string     `gorm:"column:account_agency;size:20;not null" json:"account_agency"`
	AccountNumber string     `gorm:"column:account_number;size:50;not null" json:"account_number"`
	PixToken      string     `gorm:"column:pix_token;size:255;not null" json:"pix_token"`
	PixName       string     `gorm:"column:pix_name;size:255;not null" json:"pix_name"`
	PixCity       string     `gorm:"column:pix_city;size:255;not null" json:"pix_city"`
	LogoName      string     `gorm:"column:logo_name;size:255" json:"logo_name"`
	LastRps       int64      `gorm:"column:last_rps;not null" json:"last_rps"`
	SmtpHost      string     `gorm:"column:smtp_host;size:255" json:"smtp_host"`
	SmtpPort      int        `gorm:"column:smtp_port" json:"smtp_port"`
	SmtpUser      string     `gorm:"column:smtp_user;size:255" json:"smtp_user"`
	SmtpPassword  string     `gorm:"column:smtp_password;size:255" json:"smtp_password"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
}

func (Vendor) TableName() string {
	return "vendor"
}

// Account represents a bank account owned by a vendor
type Account struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	VendorID       int64      `gorm:"column:vendor_id;not null" json:"vendor_id"`
	BankCode       string     `gorm:"column:bank_code;size:10;not null" json:"bank_code"`
	AccountNumber  string     `gorm:"column:account_number;size:50;not null" json:"account_number"`
	AccountBranch  string     `gorm:"column:account_branch;size:20;not null" json:"account_branch"`
	Status         int        `gorm:"column:status;not null;default:1" json:"status"` // 0 - inactive, 1 - active
	InitialBalance float64    `gorm:"column:initial_balance;type:numeric(10,2);not null" json:"initial_balance"`
	InitialDate    time.Time  `gorm:"column:initial_date;type:date;not null" json:"initial_date"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	// Relationships
	Vendor              *Vendor              `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
	AccountSantanderAPI *AccountSantanderAPI `gorm:"foreignKey:ID" json:"account_santander_api,omitempty"`
}

func (Account) TableName() string {
	return "bank_account"
}

// AccountSantanderAPI contains specialized API credentials for Santander bank accounts
type AccountSantanderAPI struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	ClientID     string `gorm:"column:client_id;size:50;not null" json:"client_id"`
	ClientSecret string `gorm:"column:client_secret;size:255;not null" json:"client_secret"`
}

func (AccountSantanderAPI) TableName() string {
	return "bank_account_santander_api"
}

