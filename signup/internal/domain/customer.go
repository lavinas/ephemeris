package domain

import (
	"time"

	"fmt"
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

// TableName specifies the table name for Customer model.
func (Customer) TableName() string {
	return "customer"
}

// GetByNickname retrieves a customer by their nickname.
func (c *Customer) GetByNickname(repo port.Repository, vendorID int64, nickname string) (bool, error) {
	conditions := map[string]interface{}{"nickname = ?": nickname, "vendor_id = ?": vendorID}
	resp, err := repo.Find(c, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(resp) == 0 {
		return false, nil
	}
	customer, ok := resp[0].(*Customer)
	if !ok {
		return false, fmt.Errorf("failed to cast to Customer")
	}
	*c = *customer
	return true, nil
}

// GetByDocument retrieves a customer by their document.
func (c *Customer) GetByDocument(repo port.Repository, vendorID int64, document string) (bool, error) {
	conditions := map[string]interface{}{"document = ?": document, "vendor_id = ?": vendorID}
	resp, err := repo.Find(c, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(resp) == 0 {
		return false, nil
	}
	customer, ok := resp[0].(*Customer)
	if !ok {
		return false, fmt.Errorf("failed to cast to Customer")
	}
	*c = *customer
	return true, nil
}

// Find retrieves customers based on the specified conditions.
func (c *Customer) Find(repo port.Repository) ([]port.Domain, error) {
	conditions := map[string]interface{}{}
	resp, err := repo.Find(c, conditions, 0, 0)
	if err != nil {
		return nil, err
	}
	out := make([]port.Domain, len(resp))
	for i, r := range resp {
		customer, ok := r.(*Customer)
		if !ok {
			return nil, fmt.Errorf("failed to cast to Customer")
		}
		out[i] = customer
	}
	return out, nil
}

// Save persists the customer instance to the repository.
func (c *Customer) Save(repo port.Repository) error {
	return repo.Save(c)
}
