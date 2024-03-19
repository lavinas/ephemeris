package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"slices"
	"strings"

	"github.com/klassmann/cpfcnpj"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// ErrEmptyName is a variable that represents the error message for empty name
	ErrEmptyName = "empty name"
	// ErrLongName is a variable that represents the error message for long name
	ErrLongName = "name should have at most 100"
	// ErrInvalidName is a variable that represents the error message for invalid name
	ErrInvalidName = "name should have at least two words"
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
		"phone":    c.formatPhone,
		"contact":  c.formatContact,
		"document": c.formatDocument,
	}
	for _, f := range formatMap {
		f()
	}
}

// validateName is a method that validates the name field
func (c *Client) validateName() error {
	name := strings.TrimSpace(c.Name)
	if name == "" {
		return errors.New(ErrEmptyName)
	}
	if len(name) > 100 {
		return errors.New(ErrLongName)
	}
	if len(strings.Split(name, " ")) < 2 {
		return errors.New(ErrInvalidName)
	}
	return nil
}

// formatName is a method that formats the name field
func (c *Client) formatName() {
	caser := cases.Title(language.Und)
	c.Name = caser.String(c.Name)
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
	a, _ := mail.ParseAddress(c.Email)
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
	if _, err := phonenumbers.Parse(c.Phone, "BR"); err != nil {
		return errors.New(ErrInvalidPhone)
	}
	return nil
}

// formatPhone is a method that formats the phone field
func (c *Client) formatPhone() {
	phone, _ := phonenumbers.Parse(c.Phone, "")
	c.Phone = phonenumbers.Format(phone, phonenumbers.E164)
}

// validateContact is a method that validates the contact field
func (c *Client) validateContact() error {
	contact := strings.TrimSpace(c.Contact)
	contact = strings.ToLower(contact)
	if contact == "" {
		return errors.New(ErrEmptyContact)
	}
	if !slices.Contains(ContactWays, contact) {
		return errors.New(ErrInvalidContact)
	}
	return nil
}

// formatContact is a method that formats the contact field
func (c *Client) formatContact() {
	c.Contact = strings.TrimSpace(c.Contact)
	c.Contact = strings.ToLower(c.Contact)
}

// validateDocument is a method that validates the document field
func (c *Client) validateDocument() error {
	if c.Document == "" {
		return nil
	}
	if !cpfcnpj.ValidateCPF(c.Document) && !cpfcnpj.ValidateCNPJ(c.Document) {
		return errors.New(ErrInvalidDocument)
	}
	return nil
}

// formatDocument is a method that formats the document field
func (c *Client) formatDocument() {
	cpf := cpfcnpj.NewCPF(c.Document)
	if cpf.IsValid() {
		c.Document = cpf.String()
		return
	}
	cnpj := cpfcnpj.NewCNPJ(c.Document)
	if cnpj.IsValid() {
		c.Document = cnpj.String()
		return
	}
}
