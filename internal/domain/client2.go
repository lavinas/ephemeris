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

// Client represents the client entity
type Client2 struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	Name        string    `gorm:"type:varchar(100); not null"`
	Email       string    `gorm:"type:varchar(100); not null"`
	Phone       string    `gorm:"type:varchar(20); not null"`
	Date        time.Time `gorm:"type:datetime; not null"`
	Document    string    `gorm:"type:varchar(20)"`
	Contact     string	  `gorm:"type:varchar(20)"`
}

// NewClient2 is a function that creates a new client
func NewClient2(id string, name string, email string, phone string, date time.Time, document string, contact string) *Client2 {
	return &Client2{
		ID:       id,
		Name:     name,
		Email:    email,
		Phone:    phone,
		Date:     date,
		Document: document,
		Contact:  contact,
	}
}

// Format is a method that formats the client
func (c *Client2) Format(args ...string) error {
	filled := slices.Contains(args, "filled")
	formatMap := []func(filled bool) error{
		c.formatID,
		c.formatName,
		c.formatEmail,
		c.formatPhone,
		c.formatDate,
		c.formatDocument,
		c.formatContact,
	}
	for _, f := range formatMap {
		if err := f(filled); err != nil {
			return err
		}
	}
	return nil
}

// GetID is a method that returns the id of the client
func (c *Client2) GetID() string {
	return c.ID
}

// formatID is a method that formats the id field
func (c *Client2) formatID(filled bool) error{
	id := c.formatString(c.ID)
	if id == "" {
		if filled {
			return nil
		}
		return errors.New(port.ErrEmptyID)
	}
	if len(id) > 25 {
		return errors.New(port.ErrLongID)
	}
	if len(strings.Split(id, " ")) > 1 {
		return errors.New(port.ErrInvalidID)
	}
	c.ID = strings.ToLower(id)
	return nil
}

// formatName is a method that formats the name field
func (c *Client2) formatName(filled bool) error {
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
	c.Name = cases.Title(language.Und).String(name)
	return nil
}


// formatEmail is a method that formats the email field
func (c *Client2) formatEmail(filled bool) error {
	email := c.formatString(c.Email)
	if email == "" {
		if filled {
			return nil
		}
		return errors.New(port.ErrEmptyEmail)
	}
	a, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New(port.ErrInvalidEmail)
	}
	email = a.Address
	if len(email) > 100 {
		return errors.New(port.ErrLongEmail)
	}
	c.Email = email
	return nil
}

// formatPhone is a method that formats the phone field
func (c *Client2) formatPhone(filled bool) error {
	phone := c.formatString(c.Phone)
	if phone == "" {
		if filled {
			return nil
		}
		return errors.New(port.ErrEmptyPhone)
	}
	phone = c.formatNumber(phone)
	p, err := phonenumbers.Parse(phone, "BR")
	if err != nil {
		return errors.New(port.ErrInvalidPhone)
	}
	phone = phonenumbers.Format(p, phonenumbers.E164)
	if len(phone) > 20 {
		return errors.New(port.ErrLongPhone)
	}
	c.Phone = phone
	return nil
}	

// formatDate is a method that formats the date field
func (c *Client2) formatDate(filled bool) error {
	date := c.Date
	if date.IsZero() {
		if filled {
			return nil
		}
		date = time.Now()
	}
	c.Date = date
	return nil
}

// formatDocument is a method that formats the document field
func (c *Client2) formatDocument(filled bool) error {
	document := c.formatString(c.Document)
	if document == "" {
		return nil
	}
	document = c.formatNumber(document)
	if document == "" {
		return errors.New(port.ErrInvalidDocument)
	}
	if len(document) > 20 {
		return errors.New(port.ErrLongDocument)
	}
	cpf := cpfcnpj.NewCPF(document)
	if cpf.IsValid() {
		c.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(document)
	if cnpj.IsValid() {
		c.Document = cnpj.String()
		return nil
	}
	return errors.New(port.ErrInvalidDocument)
}

// formatContact is a method that formats the contact field
func (c *Client2) formatContact(filled bool) error {
	contact := c.formatString(c.Contact)
	if contact == "" {
		if filled {
			return nil
		}
		contact = "email"
	}
	if len(contact) > 20 {
		return errors.New(port.ErrLongContact)
	}
	if !slices.Contains(ContactWays, contact) {
		return errors.New(port.ErrInvalidContact)
	}
	c.Contact = strings.ToLower(contact)
	return nil
}


// formatNumber is a method that formats a number
func (c *Client2) formatNumber(number string) string {
	re := regexp.MustCompile("[0-9]+")
	ret := ""
	for _, s := range re.FindAllString(number, -1) {
		ret += s
	}
	return ret
}

// formatString is a method that formats a string
func (c *Client2) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}