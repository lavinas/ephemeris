package domain

import (
	"time"
)

// Contract represents the contract entity
type Contract struct {
	ID          string      `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time   `gorm:"type:datetime; not null"`
	Client      *Client     `gorm:"foreignKey:ID, not null"`
	Service     *Service    `gorm:"foreignKey:ID, not null"`
	Start       time.Time   `gorm:"type:datetime; not null"`
	End         time.Time   `gorm:"type:datetime; not null"`
	Recurrence  *Recurrence `gorm:"foreignKey:ID, not null"`
	BillingType string      `gorm:"type:varchar(20), not null"`
	DueDay      int64       `gorm:"type:numeric(20), not null"`
	Price       *Price      `gorm:"foreignKey:ID, not null"`
	Bond        *Contract   `gorm:"foreignKey:ID"`
}

// NewContract creates a new contract
func NewContract(clientId string, serviceId string, start string, end string, recurrenceId string, billingType string,
	dueDay string, price string, bond string) *Contract {
	return nil
}
