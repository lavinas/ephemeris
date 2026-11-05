package domain

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Statement represents a bank account transaction entry in accounting.statement
type Statement struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID   int64      `gorm:"column:account_id;not null;index" json:"account_id"`
	SDate       *time.Time `gorm:"column:sdate;type:date;not null" json:"sdate"`
	Transaction string     `gorm:"column:transaction;size:200;not null" json:"transaction"`
	Complement  string     `gorm:"column:complement;size:200" json:"complement"`
	Document    string     `gorm:"column:document;size:100" json:"document"`
	Amount      float64    `gorm:"column:amount;type:numeric(10,2);not null" json:"amount"`
	Balance     float64    `gorm:"column:balance;type:numeric(10,2);not null" json:"balance"`
	Key         string     `gorm:"column:key;size:100;not null;uniqueIndex" json:"key"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	// Relationships
	Account               *Account               `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	StatementSantanderAPI *StatementSantanderAPI `gorm:"foreignKey:ID" json:"statement_santander_api,omitempty"`
}

func (Statement) TableName() string {
	return "statement"
}

// StatementSantanderAPI stores the specialized Santander API transaction entry in accounting.statement_santander_api
type StatementSantanderAPI struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement:false" json:"id"` // References statement(id)
	CreditDebitType    string     `gorm:"column:credit_debit_type;size:10" json:"credit_debit_type"`
	TransactionName    string     `gorm:"column:transaction_name;size:100" json:"transaction_name"`
	HistoricComplement string     `gorm:"column:historic_complement;size:100" json:"historic_complement"`
	Amount             float64    `gorm:"column:amount;type:numeric(10,2)" json:"amount"`
	TransactionDate    *time.Time `gorm:"column:transaction_date;type:date" json:"transaction_date"`
	DocumentNumber     string     `gorm:"column:document_number;size:50" json:"document_number"`
	Key                string     `gorm:"column:key;size:100" json:"key"`
}

func (StatementSantanderAPI) TableName() string {
	return "statement_santander_api"
}

// GenerateSantanderKey calculates:
// key = sha256(credit_debit_type, transaction_name, historic_complement, amount, transaction_date, document_number)
func GenerateSantanderKey(creditDebitType, transactionName, historicComplement string, amount float64, txDate *time.Time, documentNumber string) string {
	dateStr := ""
	if txDate != nil {
		dateStr = txDate.Format("2006-01-02")
	}
	raw := fmt.Sprintf("%s|%s|%s|%.2f|%s|%s", creditDebitType, transactionName, historicComplement, amount, dateStr, documentNumber)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

// GenerateStatementKey calculates:
// statement.key = sha256(account_id, occurrence, statement_santander_api.key)
func GenerateStatementKey(accountID int64, occurrence int, santanderKey string) string {
	raw := fmt.Sprintf("%d|%d|%s", accountID, occurrence, santanderKey)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}
