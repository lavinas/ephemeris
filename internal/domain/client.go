package domain

import (
	"errors"
	"net/mail"
	"slices"

	"github.com/nyaruka/phonenumbers"
	"github.com/klassmann/cpfcnpj"
)

const (
	// ErrEmptyName is a variable that represents the error message for empty name
	ErrEmptyName = "empty name"
	// ErrEmptyEmail is a variable that represents the error message for empty email
	ErrEmptyEmail = "empty email"
	// ErrInvalidEmail is a variable that represents the error message for invalid email
	ErrInvalidEmail = "invalid email"
	// ErrEmptyPhone is a variable that represents the error message for empty phone
	ErrEmptyPhone = "empty phone"
	// ErrInvalidPhone is a variable that represents the error message for invalid phone
	ErrInvalidPhone = "invalid phone"
	// ErrEmptyContact is a variable that represents the error message for empty contact
	ErrEmptyContact = "empty contact"
	// ErrInvalidContact is a variable that represents the error message for invalid contact
	ErrInvalidContact = "invalid contact"
	// ErrInvalidDocument is a variable that represents the error message for invalid document
	ErrInvalidDocument = "invalid document"
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
	message := ""
	if c.Name == "" {
		message += ErrEmptyName + ", "
	}
	if c.Email == "" {
		message += ErrEmptyEmail + ", "
	} else if _, err := mail.ParseAddress(c.Email); err != nil {
		message += ErrInvalidEmail + ", "
	}
	if c.Phone == "" {
		message += ErrEmptyPhone + ", "
	} else if _, err := phonenumbers.Parse(c.Phone, ""); err != nil {
		message += ErrInvalidPhone + ", "
	}
	if c.Contact == "" {
		message += ErrEmptyContact + ", "
	} else if !slices.Contains(ContactWays, c.Contact) {
		message += ErrInvalidContact + ", "
	}
	if c.Document != "" && !cpfcnpj.ValidateCPF(c.Document) && !cpfcnpj.ValidateCNPJ(c.Document) {
		message += ErrInvalidDocument + ", "
	}
	if message == "" {
		return nil
	}
	return errors.New(message[:len(message)-2])
}

// Validate is a method that validates the client
func (c *Client) Format() error {
	return nil
}
