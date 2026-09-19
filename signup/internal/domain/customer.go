package domain

import (
	"time"

	"signup/internal/port"
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

// GetByNickname retrieves a customer by their nickname.
func (c *Customer) GetByNickname(repo port.Repository, vendorID int64, nickname string) bool {
	resp, err := repo.Find(1, 1, map[string]interface{}{"nickname = ?": nickname, "vendor_id = ?": vendorID})
	if err != nil || len(resp) == 0 {
		return false
	}
	customer, ok := resp[0].(*Customer)
	if !ok {
		return false
	}
	*c = *customer
	return true
}

// GetByDocument retrieves a customer by their document.
func (c *Customer) GetByDocument(repo port.Repository, vendorID int64, document string) bool {
	resp, err := repo.Find(1, 1, map[string]interface{}{"document = ?": document, "vendor_id = ?": vendorID})
	if err != nil || len(resp) == 0 {
		return false
	}
	customer, ok := resp[0].(*Customer)
	if !ok {
		return false
	}
	*c = *customer
	return true
}

// TableName specifies the table name for Customer model.
func (Customer) TableName() string {
	return "customer"
}
