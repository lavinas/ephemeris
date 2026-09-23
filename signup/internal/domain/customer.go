package domain

import (
	"time"

	"fmt"
	"signup/internal/port"
)

type Customer struct {
	DomainBase
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	VendorID  int64     `gorm:"not null;index"`
	Name      string    `gorm:"not null"`
	Nickname  string    `gorm:"not null;unique"`
	Document  *string   `gorm:"unique"`
	Email     *string   `gorm:"null"`
	Whatsapp  *string   `gorm:"null"`
	Status    *int      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// NewCustomer creates a new Customer instance with the provided details.
func NewCustomer(repo port.Repository, vendorID int64, name, nickname, document, email, whatsapp *string) *Customer {
	status := 1
	return &Customer{
		DomainBase: DomainBase{Repo: repo},
		ID:         0,
		VendorID:   vendorID,
		Name:       *name,
		Nickname:   *nickname,
		Status:     &status,
		Document:   document,
		Email:      email,
		Whatsapp:   whatsapp,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// StartCustomer represents the input data required to create a new customer.
func StartCustomer(repo port.Repository) *Customer {
	return &Customer{DomainBase: DomainBase{Repo: repo}}
}

// TableName specifies the table name for Customer model.
func (Customer) TableName() string {
	return "customer"
}

// GetByNickname retrieves a customer by their nickname.
func (c *Customer) GetByNickname(vendorID int64, nickname string) (bool, error) {
	conditions := map[string]interface{}{"nickname = ?": nickname, "vendor_id = ?": vendorID}
	resp, err := c.Repo.Find(c, conditions, 1, 1)
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
func (c *Customer) GetByDocument(vendorID int64, document string) (bool, error) {
	conditions := map[string]interface{}{"document = ?": document, "vendor_id = ?": vendorID}
	resp, err := c.Repo.Find(c, conditions, 1, 1)
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
func (c *Customer) Find(page, pageSize int) ([]port.Domain, error) {
	conditions := map[string]interface{}{}
	if c.VendorID > 0 {
		conditions["vendor_id = ?"] = c.VendorID
	}
	if c.Name != "" {
		conditions["name like ?"] = "%" + c.Name + "%"
	}
	if c.Nickname != "" {
		conditions["nickname like ?"] = "%" + c.Nickname + "%"
	}
	if c.Document != nil {
		conditions["document like ?"] = "%" + *c.Document + "%"
	}
	if c.Email != nil {
		conditions["email like ?"] = "%" + *c.Email + "%"
	}
	if c.Whatsapp != nil {
		conditions["whatsapp like ?"] = "%" + *c.Whatsapp + "%"
	}
	if c.Status != nil {
		conditions["status = ?"] = *c.Status
	}
	resp, err := c.Repo.Find(c, conditions, page, pageSize)
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

// Validate validates customer data before saving.
func (c *Customer) Validate() error {
	if c.VendorID == 0 {
		return fmt.Errorf("vendor_id is required")
	}
	vendor := StartVendor(c.Repo)
	if ok, err := vendor.GetByID(c.VendorID); err != nil || !ok {
		return fmt.Errorf("vendor not found")
	}
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	if c.Nickname != "" {
		if err := c.ValidateNickname(c.Nickname); err != nil {
			return fmt.Errorf("nickname is invalid")
		}
	}
	if c.Document != nil {
		if newDoc, err := c.ValidateCpfCnpj(*c.Document); err != nil {
			return fmt.Errorf("document is invalid")
		} else {
			c.Document = &newDoc
		}
	}
	if c.Email != nil {
		if err := c.ValidateEmail(*c.Email); err != nil {
			return fmt.Errorf("email is invalid")
		}
	}
	if c.Whatsapp != nil {
		if _, err := c.ValidateCellNumber(*c.Whatsapp); err != nil {
			return fmt.Errorf("whatsapp is invalid")
		}
	}
	if c.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}
	if c.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	if c.Status == nil || (*c.Status != 1 && *c.Status != 0) {
		return fmt.Errorf("status is required or invalid")
	}
	return nil
}

// Save persists the customer instance to the repository.
func (c *Customer) Save() error {
	if err := c.Validate(); err != nil {
		return err
	}
	return c.Repo.Save(c)
}
