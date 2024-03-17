package domain

import (
	// "github.com/klassmann/cpfcnpj"
)

var (
	ContactWays = []string{"email", "whatsapp"}
)

// Client represents the client entity
type Client struct {
	Base     `gorm:"embedded"`
	Name     string `gorm:"type:varchar(100), not null"`
	Email    string `gorm:"type:varchar(100), not null"`
	Phone    string `gorm:"type:varchar(20), not null"`
	Contact  string `gorm:"type:varchar(20), not null"`
	Document string `gorm:"type:varchar(20)"`
}

// NewClient is a function that creates a new client
func NewClient(id string, name string, email string, phone string, contact string, document string) *Client {
	return &Client{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Contact:  contact,
		Document: document,
	}
}

// Validate is a method that validates the client
func (c *Client) Validate() error {
	return nil
}

// Validate is a method that validates the client
func (c *Client) Format() error {
	return nil
}
