package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"slices"

	"github.com/klassmann/cpfcnpj"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	ContactWays = []string{"email", "whatsapp"}
)

// Client represents the client entity
type Client struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Name        string    `gorm:"type:varchar(100); not null"`
	Responsible string    `gorm:"type:varchar(100)"`
	Email       string    `gorm:"type:varchar(100); not null"`
	Phone       string    `gorm:"type:varchar(20); not null"`
	Contact     string    `gorm:"type:varchar(20); not null"`
	Document    string    `gorm:"type:varchar(20)"`
}

// NewClient is a function that creates a new client
func NewClient(id string, name string, responsible string, email string, phone string, contact string, document string) *Client {
	return &Client{
		ID:          id,
		CreatedAt:   time.Now(),
		Name:        name,
		Responsible: responsible,
		Email:       email,
		Phone:       phone,
		Contact:     contact,
		Document:    document,
	}
}

// Validate is a method that validates the client
func (c *Client) Validate(args ...string) error {
	filled := slices.Contains(args, "filled")
	message := ""
	validSlice := []func(filled bool) error{
		c.validateID,
		c.validateName,
		c.validateResponsible,
		c.validateEmail,
		c.validatePhone,
		c.validateContact,
		c.validateDocument,
	}
	for _, f := range validSlice {
		if err := f(filled); err != nil {
			message += err.Error() + ", "
		}
	}
	if message == "" {
		return nil
	}
	return errors.New(message[:len(message)-2])
}

// Format is a method that formats the client
func (c *Client) Format(args ...string) error {
	filled := slices.Contains(args, "filled")
	formatMap := []func(filled bool) error{
		c.formatID,
		c.formatName,
		c.formatResponsible,
		c.formatEmail,
		c.formatPhone,
		c.formatContact,
		c.formatDocument,
	}
	for _, f := range formatMap {
		if err := f(filled); err != nil {
			return err
		}
	}
	return nil
}

// GetID is a method that returns the id of the client
func (c *Client) GetID() string {
	return c.ID
}

// validateID is a method that validates the id field
func (c *Client) validateID(filled bool) error {
	if filled && c.ID == "" {
		return nil
	}
	if c.ID == "" {
		return errors.New(port.ErrEmptyID)
	}
	if len(strings.Split(c.ID, " ")) > 1 {
		return errors.New(port.ErrInvalidID)
	}
	return nil
}

// formatID is a method that formats the id field
func (c *Client) formatID(filled bool) error{
	id := c.formatString(c.ID)
	if id == "" {
		if filled {
			return nil
		}
		return errors.New(port.ErrEmptyID)
	}
	if len(c.ID) > 25 {
		return errors.New(port.ErrLongID)
	}
	if len(strings.Split(c.ID, " ")) > 1 {
		return errors.New(port.ErrInvalidID)
	}
	c.ID = strings.ToLower(c.ID)
	return nil
}

// validateName is a method that validates the name field
func (c *Client) validateName(filled bool) error {
	if filled && c.Name == "" {
		return nil
	}
	name := strings.TrimSpace(c.Name)
	if name == "" {
		return errors.New(port.ErrEmptyName)
	}
	if len(name) > 100 {
		return errors.New(port.ErrLongName)
	}
	if len(strings.Split(name, " ")) < 2 {
		return errors.New(port.ErrInvalidName)
	}
	return nil
}

// formatName is a method that formats the name field
func (c *Client) formatName(filled bool) error {
	name := c.formatString(c.Name)
	if name == "" {
		if filled {
			return nil
		}
		return errors.New(port.ErrEmptyName)
	}
	if len(name) > 100 {
		return errors.New(port.ErrLongName)
	}
	if len(strings.Split(name, " ")) < 2 {
		return errors.New(port.ErrInvalidName)
	}
	caser := cases.Title(language.Und)
	c.Name = caser.String(c.Name)
	return nil
}

// validateResponsible is a method that validates the responsible field
func (c *Client) validateResponsible(filled bool) error {
	responsible := strings.TrimSpace(c.Responsible)
	if responsible == "" {
		return nil
	}
	if len(responsible) > 100 {
		return errors.New(port.ErrLongResponsible)
	}
	if len(strings.Split(responsible, " ")) < 2 {
		return errors.New(port.ErrInvalidResponsible)
	}
	return nil
}

// formatResponsible is a method that formats the responsible field
func (c *Client) formatResponsible(filled bool) error{
	responsible := c.formatString(c.Responsible)
	if responsible == "" {
		return nil
	}
	if len(responsible) > 100 {
		return errors.New(port.ErrLongResponsible)
	}
	


	if filled && c.Responsible == "" {
		return nil
	}
	if err := c.validateResponsible(filled); err != nil {
		c.Responsible = ""
		return err
	}
	caser := cases.Title(language.Und)
	c.Responsible = caser.String(c.Responsible)
	c.Responsible = strings.TrimSpace(c.Responsible)
	space := regexp.MustCompile(`\s+`)
	c.Responsible = space.ReplaceAllString(c.Responsible, " ")
}

// validateEmail is a method that validates the email field
func (c *Client) validateEmail(filled bool) error {
	if filled && c.Email == "" {
		return nil
	}
	if c.Email == "" {
		return errors.New(port.ErrEmptyEmail)
	}
	if len(c.Email) > 100 {
		return errors.New(port.ErrLongEmail)
	}
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return errors.New(port.ErrInvalidEmail)
	}
	return nil
}

// formatEmail is a method that formats the email field
func (c *Client) formatEmail(filled bool) error {
	if filled && c.Email == "" {
		return nil
	}
	if err := c.validateEmail(filled); err != nil {
		c.Email = ""
		return err
	}
	a, _ := mail.ParseAddress(c.Email)
	c.Email = a.Address
}

// validatePhone is a method that validates the phone field
func (c *Client) validatePhone(filled bool) error {
	if filled && c.Phone == "" {
		return nil
	}
	if c.Phone == "" {
		return errors.New(port.ErrEmptyPhone)
	}
	if len(c.Phone) > 20 {
		return errors.New(port.ErrLongPhone)
	}
	p, err := phonenumbers.Parse(c.Phone, "BR")
	if err != nil {
		return errors.New(port.ErrInvalidPhone)
	}
	if !phonenumbers.IsValidNumberForRegion(p, "BR") {
		return errors.New(port.ErrInvalidPhone)
	}
	return nil
}

// formatPhone is a method that formats the phone field
func (c *Client) formatPhone(filled bool) error {
	phone := c.formatString(c.Phone)
	if filled && c.Phone == "" {
		return nil
	}
	if err := c.validatePhone(filled); err != nil {
		c.Phone = ""
		return err
	}
	phone = c.formatNumber(phone)
	phone, _ = phonenumbers.Parse(phone, "BR")
	c.Phone = phonenumbers.Format(phone, phonenumbers.E164)
}

// validateContact is a method that validates the contact field
func (c *Client) validateContact(filled bool) error {
	contact := c.formatString(c.Contact)
	if filled && c.Contact == "" {
		return nil
	}
	contact = strings.ToLower(contact)
	if contact == "" {
		return errors.New(port.ErrEmptyContact)
	}
	if len(contact) > 20 {
		return errors.New(port.ErrLongContact)
	}
	if !slices.Contains(ContactWays, contact) {
		return errors.New(port.ErrInvalidContact)
	}
	return nil
}

// formatContact is a method that formats the contact field
func (c *Client) formatContact(filled bool) error {
	contact := c.formatString(c.Contact)	
	if filled && c.Contact == "" {
		return nil
	}
	if err := c.validateContact(filled); err != nil {
		c.Contact = ""
		return err
	}
	contact = strings.ToLower(contact)
	if !slices.Contains(ContactWays, contact) {
		c.Contact = ""
	}
	c.Contact = contact
}

// validateDocument is a method that validates the document field
func (c *Client) validateDocument(filled bool) error {
	document := c.formatString(c.Document)
	if document == "" {
		return nil
	}
	document = c.formatNumber(c.Document)
	if document == "" {
		return errors.New(port.ErrInvalidDocument)
	}
	if len(document) > 20 {
		return errors.New(port.ErrLongDocument)
	}
	if !cpfcnpj.ValidateCPF(document) && !cpfcnpj.ValidateCNPJ(document) {
		return errors.New(port.ErrInvalidDocument)
	}
	return nil
}

// formatDocument is a method that formats the document field
func (c *Client) formatDocument(filled bool) error {
	doc := c.formatString(c.Document)
	if doc == "" {
		return nil
	}
	if err := c.validateDocument(filled); err != nil {
		return err
	}
	document := c.formatNumber(c.Document)
	re := regexp.MustCompile("[0-9]+")
	document = re.FindString(document)
	cpf := cpfcnpj.NewCPF(document)
	if cpf.IsValid() {
		c.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(document)
	c.Document = cnpj.String()
	return nil
}

// formatNumber is a method that formats a number
func (c *Client) formatNumber(number string) string {
	re := regexp.MustCompile("[0-9]+")
	ret := ""
	for _, s := range re.FindAllString(number, -1) {
		ret += s
	}
	return ret
}

// formatString is a method that formats a string
func (c *Client) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}