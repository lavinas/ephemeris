package domain

import (
	"time"
)

// Billing represents the billing type entity
type ContractBillingType struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

// Contract represents the contract entity
type Contract struct {
	Base        `gorm:"embedded"`
	Client      *Client              `gorm:"foreignKey:ID, not null"`
	Service     *Service             `gorm:"foreignKey:ID, not null"`
	Start       time.Time            `gorm:"type:datetime; not null"`
	End         time.Time            `gorm:"type:datetime; not null"`
	Recurrence  *Recurrence          `gorm:"foreignKey:ID, not null"`
	BillingType *ContractBillingType `gorm:"foreignKey:ID, not null"`
	DueDay      int64                `gorm:"type:numeric(20), not null"`
	Price       *Price               `gorm:"foreignKey:ID, not null"`
	Bond        *Contract            `gorm:"foreignKey:ID"`
}
