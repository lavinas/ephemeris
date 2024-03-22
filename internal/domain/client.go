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
	// ErrLongResponsible is a variable that represents the error message for long responsible
	ErrLongResponsible = "responsible should have at most 100"
	// ErrInvalidResponsible is a variable that represents the error message for invalid responsible
	ErrInvalidResponsible = "responsible should have at least two words"
	// ErrEmptyEmail is a variable that represents the error message for empty email
	ErrEmptyEmail = "empty email"
	// ErrInvalidEmail is a variable that represents the error message for invalid email
	ErrInvalidEmail = "invalid email"
	// ErrLongEmail is a variable that represents the error message for long email
	ErrLongEmail = "email should have at most 100"
	// ErrEmptyPhone is a variable that represents the error message for empty phone
	ErrEmptyPhone = "empty phone"
	// ErrLongPhone is a variable that represents the error message for long phone
	ErrLongPhone = "phone should have at most 20"
	// ErrInvalidPhone is a variable that represents the error message for invalid phone
	ErrInvalidPhone = "invalid phone"
	// ErrEmptyContact is a variable that represents the error message for empty contact
	ErrEmptyContact = "empty contact"
	// ErrLongContract is a variable that represents the error message for long contact
	ErrLongContact = "contact should have at most 20"
	// ErrInvalidContact is a variable that represents the error message for invalid contact
	ErrInvalidContact = "invalid contact"
	// ErrInvalidDocument is a variable that represents the error message for invalid document
	ErrInvalidDocument = "invalid document"
	// ErrLongDocument is a variable that represents the error message for long document
	ErrLongDocument = "document should have at most 20"
)

var (
	ContactWays = []string{"email", "whatsapp"}
)

// Client represents the client entity
type Client struct {
	Base
	Name        string `gorm:"type:varchar(100); not null"`
	Responsible string `gorm:"type:varchar(100)"`
	Email       string `gorm:"type:varchar(100); not null"`
	Phone       string `gorm:"type:varchar(20); not null"`
	Contact     string `gorm:"type:varchar(20); not null"`
	Document    string `gorm:"type:varchar(20)"`
}

// NewClient is a function that creates a new client
func NewClient(id string, name string, responsible string, email string, phone string, contact string, document string) *Client {
	return &Client{
		Base:        *NewBase(id),
		Name:        name,
		Responsible: responsible,
		Email:       email,
		Phone:       phone,
		Contact:     contact,
		Document:    document,
	}
}

// Validate is a method that validates the client
func (c *Client) Validate() error {
	message := ""
	validSlice := []func() error{
		c.validateName,
		c.validateResponsible,
		c.validateEmail,
		c.validatePhone,
		c.validateContact,
		c.validateDocument,
	}
	for _, f := range validSlice {
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
func (c *Client) Format() error {
	formatMap := map[string]func() error{
		"name":        c.formatName,
		"responsible": c.formatResponsible,
		"email":       c.formatEmail,
		"phone":       c.formatPhone,
		"contact":     c.formatContact,
		"document":    c.formatDocument,
	}
	for _, f := range formatMap {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

// String is a method that returns a string representation of the client
func (c *Client) String() string {
	ret := ""
	ret += "id: " + c.ID + "; "
	ret += "name: " + c.Name + "; "
	if c.Responsible != "" {
		ret += "responsible: " + c.Responsible + "; "
	}
	ret += "email: " + c.Email + "; "
	ret += "phone: " + c.Phone + "; "
	ret += "contact: " + c.Contact
	if c.Document != "" {
		ret += "; document: " + c.Document
	}
	return ret
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
func (c *Client) formatName() error {
	if err := c.validateName(); err != nil {
		return err
	}
	caser := cases.Title(language.Und)
	c.Name = caser.String(c.Name)
	c.Name = strings.TrimSpace(c.Name)
	space := regexp.MustCompile(`\s+`)
	c.Name = space.ReplaceAllString(c.Name, " ")
	return nil
}

// validateResponsible is a method that validates the responsible field
func (c *Client) validateResponsible() error {
	responsible := strings.TrimSpace(c.Responsible)
	if responsible == "" {
		return nil
	}
	if len(responsible) > 100 {
		return errors.New(ErrLongResponsible)
	}
	if len(strings.Split(responsible, " ")) < 2 {
		return errors.New(ErrInvalidResponsible)
	}
	return nil
}

// formatResponsible is a method that formats the responsible field
func (c *Client) formatResponsible() error {
	if err := c.validateResponsible(); err != nil {
		return err
	}
	caser := cases.Title(language.Und)
	c.Responsible = caser.String(c.Responsible)
	c.Responsible = strings.TrimSpace(c.Responsible)
	space := regexp.MustCompile(`\s+`)
	c.Responsible = space.ReplaceAllString(c.Responsible, " ")
	return nil
}

// validateEmail is a method that validates the email field
func (c *Client) validateEmail() error {
	if c.Email == "" {
		return errors.New(ErrEmptyEmail)
	}
	if len(c.Email) > 100 {
		return errors.New(ErrLongEmail)
	}
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return errors.New(ErrInvalidEmail)
	}
	return nil
}

// formatEmail is a method that formats the email field
func (c *Client) formatEmail() error {
	if err := c.validateEmail(); err != nil {
		return err
	}
	a, _ := mail.ParseAddress(c.Email)
	c.Email = a.Address
	return nil
}

// validatePhone is a method that validates the phone field
func (c *Client) validatePhone() error {
	if c.Phone == "" {
		return errors.New(ErrEmptyPhone)
	}
	if len(c.Phone) > 20 {
		return errors.New(ErrLongPhone)
	}
	p, err := phonenumbers.Parse(c.Phone, "BR")
	if err != nil {
		return errors.New(ErrInvalidPhone)
	}
	if !phonenumbers.IsValidNumberForRegion(p, "BR") {
		return errors.New(ErrInvalidPhone)
	}
	return nil
}

// formatPhone is a method that formats the phone field
func (c *Client) formatPhone() error {
	if err := c.validatePhone(); err != nil {
		return err
	}
	phone, _ := phonenumbers.Parse(c.Phone, "BR")
	c.Phone = phonenumbers.Format(phone, phonenumbers.E164)
	return nil
}

// validateContact is a method that validates the contact field
func (c *Client) validateContact() error {
	contact := strings.TrimSpace(c.Contact)
	contact = strings.ToLower(contact)
	if contact == "" {
		return errors.New(ErrEmptyContact)
	}
	if len(contact) > 20 {
		return errors.New(ErrLongContact)
	}
	if !slices.Contains(ContactWays, contact) {
		return errors.New(ErrInvalidContact)
	}
	return nil
}

// formatContact is a method that formats the contact field
func (c *Client) formatContact() error {
	if err := c.validateContact(); err != nil {
		return err
	}
	contact := strings.TrimSpace(c.Contact)
	contact = strings.ToLower(contact)
	if !slices.Contains(ContactWays, contact) {
		c.Contact = ""
	}
	c.Contact = contact
	return nil
}

// validateDocument is a method that validates the document field
func (c *Client) validateDocument() error {
	document := strings.TrimSpace(c.Document)
	if document == "" {
		return nil
	}
	if len(document) > 20 {
		return errors.New(ErrLongDocument)
	}
	if !cpfcnpj.ValidateCPF(document) && !cpfcnpj.ValidateCNPJ(document) {
		return errors.New(ErrInvalidDocument)
	}
	return nil
}

// formatDocument is a method that formats the document field
func (c *Client) formatDocument() error {
	if err := c.validateDocument(); err != nil {
		return err
	}
	document := strings.TrimSpace(c.Document)
	cpf := cpfcnpj.NewCPF(document)
	if cpf.IsValid() {
		c.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(document)
	c.Document = cnpj.String()
	return nil
}
