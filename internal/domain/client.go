package domain

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/klassmann/cpfcnpj"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/lavinas/ephemeris/pkg"
)

var (
	// ContactWays is a slice that contains the ways to contact a client
	ContactWays = []string{pkg.ContactEmail, pkg.ContactWhatsapp, pkg.ContactAll}
)

// Client represents the client entity
type Client struct {
	ID       string    `gorm:"type:varchar(25); primaryKey"`
	Date     time.Time `gorm:"type:datetime; not null; index"`
	Name     string    `gorm:"type:varchar(100); not null; index"`
	Email    string    `gorm:"type:varchar(100);  not null; index"`
	Phone    string    `gorm:"type:varchar(20); not null; index"`
	Contact  string    `gorm:"type:varchar(20); not null; index"`
	Document *string   `gorm:"type:varchar(20); null; index"`
}

// NewClient is a function that creates a new client
func NewClient(id, date, name, email, phone, document, contact string) *Client {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	var doc *string = nil
	if document != "" {
		doc = &document
	}
	return &Client{
		ID:       id,
		Date:     fdate,
		Name:     name,
		Email:    email,
		Phone:    phone,
		Document: doc,
		Contact:  contact,
	}
}

// Format is a method that formats the client
func (c *Client) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	formatMap := []func(filled bool) error{
		c.formatID,
		c.formatDate,
		c.formatName,
		c.formatEmail,
		c.formatPhone,
		c.formatDocument,
		c.formatContact,
	}
	message := ""
	for _, f := range formatMap {
		if err := f(filled); err != nil {
			message += err.Error() + " | "
		}
	}
	if err := c.validateDuplicity(repo, noduplicity); err != nil {
		message += err.Error() + " | "
	}
	if message != "" {
		return errors.New(message[:len(message)-3])
	}
	return nil
}

// Exists is a function that checks if a client exists
func (c *Client) Load(repo port.Repository) (bool, error) {
	return repo.Get(c, c.ID)
}

// GetID is a method that returns the id of the client
func (c *Client) GetID() string {
	return c.ID
}

// Get is a method that returns the client
func (c *Client) Get() port.Domain {
	return c
}

// GetEmpty is a method that returns an empty client with just id
func (c *Client) GetEmpty() port.Domain {
	return &Client{}
}

// TableName returns the table name for database
func (b *Client) TableName() string {
	return "client"
}

// formatID is a method that formats the id field
func (c *Client) formatID(filled bool) error {
	id := c.formatString(c.ID)
	if id == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	if len(id) > 25 {
		return errors.New(pkg.ErrLongID)
	}
	if len(strings.Split(id, " ")) > 1 {
		return errors.New(pkg.ErrInvalidID)
	}
	c.ID = strings.ToLower(id)
	return nil
}

// formatDate is a method that formats the date field
func (c *Client) formatDate(filled bool) error {
	if c.Date.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
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
		return errors.New(pkg.ErrEmptyName)
	}
	if len(name) > 100 {
		return errors.New(pkg.ErrLongName)
	}
	if len(strings.Split(name, " ")) < 2 {
		return errors.New(pkg.ErrInvalidName)
	}
	c.Name = cases.Title(language.Und).String(name)
	return nil
}

// formatEmail is a method that formats the email field
func (c *Client) formatEmail(filled bool) error {
	email := c.formatString(c.Email)
	if email == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyEmail)
	}
	a, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New(pkg.ErrInvalidEmail)
	}
	email = a.Address
	if len(email) > 100 {
		return errors.New(pkg.ErrLongEmail)
	}
	c.Email = email
	return nil
}

// formatPhone is a method that formats the phone field
func (c *Client) formatPhone(filled bool) error {
	phone := c.formatString(c.Phone)
	if phone == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyPhone)
	}
	phone = c.formatNumber(phone)
	p, err := phonenumbers.Parse(phone, "BR")
	if err != nil {
		return errors.New(pkg.ErrInvalidPhone)
	}
	phone = phonenumbers.Format(p, phonenumbers.E164)
	if len(phone) > 20 {
		return errors.New(pkg.ErrLongPhone)
	}
	c.Phone = phone
	return nil
}

// formatDocument is a method that formats the document field
func (c *Client) formatDocument(filled bool) error {
	if c.Document == nil {
		return nil
	}
	if *c.Document = c.formatNumber(*c.Document); *c.Document == "" {
		return nil
	}
	cpf := cpfcnpj.NewCPF(*c.Document)
	if cpf.IsValid() {
		*c.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(*c.Document)
	if cnpj.IsValid() {
		*c.Document = cnpj.String()
		return nil
	}
	return errors.New(pkg.ErrInvalidDocument)
}

// formatContact is a method that formats the contact field
func (c *Client) formatContact(filled bool) error {
	contact := c.formatString(c.Contact)
	if contact == "" {
		if filled {
			return nil
		}
		c.Contact = contact
		return nil
	}
	if len(contact) > 20 {
		return errors.New(pkg.ErrLongContact)
	}
	if !slices.Contains(ContactWays, contact) {
		ways := strings.Join(ContactWays, ", ")
		ways = ways[:len(ways)-2]
		return fmt.Errorf(pkg.ErrInvalidContact, ways)
	}
	c.Contact = strings.ToLower(contact)
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

// validateDuplicity is a method that validates the duplicity of a client
func (c *Client) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Client{}, c.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}
