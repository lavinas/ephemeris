package domain

import (
	"errors"
	"net/mail"
	"slices"
	"strings"
	"regexp"

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
	validMap := map[string]func() error{
		"name":     c.validateName,
		"email":    c.validateEmail,
		"phone":    c.validatePhone,
		"contact":  c.validateContact,
		"document": c.validateDocument,
	}
	for _, f := range validMap {
		if err := f(); err != nil {
			message += err.Error() + ", "
		}
	}
	if message == "" {
		return nil
	}
	return errors.New(message[:len(message)-2])
}

// Format is a method that formats the client
func (c *Client) Format() {
	formatMap := map[string]func(){
		"name":     c.formatName,
		"email":    c.formatEmail,
		// "phone":    c.formatPhone,
		// "contact":  c.formatContact,
		// "document": c.formatDocument,
	}
	for _, f := range formatMap {
		f()
	}
}

// validateName is a method that validates the name field
func (c *Client) validateName() error {
	if c.Name == "" {
		return errors.New(ErrEmptyName)
	}
	return nil
}

// formatName is a method that formats the name field
func (c *Client) formatName() {
	c.Name = strings.Title(strings.ToLower(c.Name))
	c.Name = strings.TrimSpace(c.Name)
	space := regexp.MustCompile(`\s+`)
	c.Name = space.ReplaceAllString(c.Name, " ")
}

// validateEmail is a method that validates the email field
func (c *Client) validateEmail() error {
	if c.Email == "" {
		return errors.New(ErrEmptyEmail)
	}
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return errors.New(ErrInvalidEmail)
	}
	return nil
}

// formatEmail is a method that formats the email field
func (c *Client) formatEmail() {
	a, _:= mail.ParseAddress(c.Email)
	if a == nil {
		c.Email = ""
		return
	}
	c.Email = a.Address
}
	
// validatePhone is a method that validates the phone field
func (c *Client) validatePhone() error {
	if c.Phone == "" {
		return errors.New(ErrEmptyPhone)
	}
	if _, err := phonenumbers.Parse(c.Phone, ""); err != nil {
		return errors.New(ErrInvalidPhone)
	}
	return nil
}

// validateContact is a method that validates the contact field
func (c *Client) validateContact() error {
	if c.Contact == "" {
		return errors.New(ErrEmptyContact)
	}
	if !slices.Contains(ContactWays, c.Contact) {
		return errors.New(ErrInvalidContact)
	}
	return nil
}

// validateDocument is a method that validates the document field
func (c *Client) validateDocument() error {
	if c.Document == ""{
		return nil
	}
	if !cpfcnpj.ValidateCPF(c.Document) && !cpfcnpj.ValidateCNPJ(c.Document) {
		return errors.New(ErrInvalidDocument)
	}
	return nil
}

