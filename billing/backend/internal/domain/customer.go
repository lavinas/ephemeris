package domain

import (
	"time"
)

type Customer struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	VendorID  int64     `gorm:"not null;index"`
	Name      string    `gorm:"not null"`
	Nickname  string    `gorm:"not null;unique"`
	Document  *string   `gorm:"unique"`
	Email     *string   `gorm:"null"`
	Whatsapp  *string   `gorm:"null"`
	Status    int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// NewCustomer creates a new Customer instance with the provided details.
func NewCustomer(vendorID int64, name, nickname, document, email, whatsapp *string) *Customer {
	return &Customer{
		ID:        0,
		VendorID:  vendorID,
		Name:      *name,
		Nickname:  *nickname,
		Status:    1,
		Document:  document,
		Email:     email,
		Whatsapp:  whatsapp,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TableName specifies the table name for Customer model.
func (Customer) TableName() string {
	return "customer"
}
