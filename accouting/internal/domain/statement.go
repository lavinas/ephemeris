package domain

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Statement represents generic bank account transaction entry
type Statement struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID int64      `gorm:"column:account_id;not null;index" json:"account_id"`
	SDate     *time.Time `gorm:"column:sdate;type:date" json:"sdate"`
	Amount    float64    `gorm:"column:amount;type:numeric(10,2)" json:"amount"`
	Memo      string     `gorm:"column:memo;size:200" json:"memo"`
	Partie    string     `gorm:"column:partie;size:50" json:"partie"`
	Checksum  string     `gorm:"column:checksum;size:100;uniqueIndex" json:"checksum"`
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	// Relationships
	Account              *Account              `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	StatementSantanderAPI *StatementSantanderAPI `gorm:"foreignKey:ID" json:"statement_santander_api,omitempty"`
}

func (Statement) TableName() string {
	return "statement"
}

// GenerateChecksum calculates a SHA-256 hash based on (account_id, sdate, amount, memo, partie)
func GenerateChecksum(accountID int64, sDate *time.Time, amount float64, memo, partie string) string {
	dateStr := ""
	if sDate != nil {
		dateStr = sDate.Format("2006-01-02")
	}
	// Formatted to two decimal places for stable amount representation
	raw := fmt.Sprintf("%d|%s|%.2f|%s|%s", accountID, dateStr, amount, memo, partie)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

// StatementSantanderAPI stores the original raw data received from Santander API
type StatementSantanderAPI struct {
	ID                 int64      `gorm:"primaryKey" json:"id"` // References statement(id)
	TransactionID      string     `gorm:"column:transaction_id;size:100" json:"transaction_id"`
	TransactionStatus  string     `gorm:"column:transaction_status;size:100" json:"transaction_status"`
	TransactionSign    string     `gorm:"column:transaction_sign;size:1" json:"transaction_sign"` // + or -
	TransactionName    string     `gorm:"column:transaction_name;size:200" json:"transaction_name"`
	TransactionType    string     `gorm:"column:transaction_type;size:100" json:"transaction_type"`
	TransactionAmount  float64    `gorm:"column:transaction_amount;type:numeric(10,2)" json:"transaction_amount"`
	TransactionCurrency string    `gorm:"column:transaction_currency;size:3" json:"transaction_currency"`
	TransactionDate    *time.Time `gorm:"column:transaction_date;type:date" json:"transaction_date"`
	PartieDocument     string     `gorm:"column:partie_document;size:50" json:"partie_document"`
	PartieType         string     `gorm:"column:partie_type;size:100" json:"partie_type"`
	PartieBankCode     string     `gorm:"column:partie_bank_code;size:10" json:"partie_bank_code"`
	PartieBranchCode   string     `gorm:"column:partie_branch_code;size:10" json:"partie_branch_code"`
	PartieAccount      string     `gorm:"column:partie_account;size:50" json:"partie_account"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

func (StatementSantanderAPI) TableName() string {
	return "statement_santander_api"
}

