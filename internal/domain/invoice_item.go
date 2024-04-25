package domain

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// InvoiceItem represents the invoice item entity
type InvoiceItem struct {
	ID          string  `gorm:"type:varchar(25); primaryKey"`
	InvoiceID   string  `gorm:"type:varchar(25); not null"`
	AgendaID    *string `gorm:"type:varchar(25); null"`
	Value       float64 `gorm:"type:numeric(20,2); not null"`
	Description string  `gorm:"type:varchar(100); not null"`
}

// NewInvoiceItem creates a new invoice item domain entity
func NewInvoiceItem(id, invoiceID, agendaID, value, description string) *InvoiceItem {
	invoiceItem := &InvoiceItem{}
	invoiceItem.ID = id
	invoiceItem.InvoiceID = invoiceID
	invoiceItem.AgendaID = nil
	if agendaID != "" {
		invoiceItem.AgendaID = &agendaID
	}
	var err error
	if invoiceItem.Value, err = strconv.ParseFloat(value, 64); err != nil {
		invoiceItem.Value = math.NaN()
	}
	invoiceItem.Description = description
	return invoiceItem
}

// Format formats the invoice item
func (i *InvoiceItem) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := i.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatInvoiceID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatAgendaID(repo); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatValue(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatDescription(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.validateDuplicity(repo, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a function that checks if a client exists
func (c *InvoiceItem) Load(repo port.Repository) (bool, error) {
	return repo.Get(c, c.ID)
}

// GetID is a method that returns the id of the client
func (c *InvoiceItem) GetID() string {
	return c.ID
}

// Get is a method that returns the client
func (c *InvoiceItem) Get() port.Domain {
	return c
}

// GetEmpty is a method that returns an empty client with just id
func (c *InvoiceItem) GetEmpty() port.Domain {
	return &InvoiceItem{}
}

// TableName returns the table name for database
func (b *InvoiceItem) TableName() string {
	return "invoice_item"
}

// formatID is a method that formats the id of the contract
func (c *InvoiceItem) formatID(filled bool) error {
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

// formatInvoiceID is a method that formats the invoice id
func (c *InvoiceItem) formatInvoiceID(repo port.Repository, filled bool) error {
	c.InvoiceID = c.formatString(c.InvoiceID)
	if c.InvoiceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyInvoice)
	}
	invoice := Invoice{ID: c.InvoiceID}
	if ok, err := invoice.Load(repo); err != nil {
		return err
	} else if !ok {
		return errors.New(pkg.ErrInvoiceNotFound)
	}
	return nil
}

// formatAgendaID is a method that formats the agenda id
func (c *InvoiceItem) formatAgendaID(repo port.Repository) error {
	if c.AgendaID == nil {
		return nil
	}
	agenda := Agenda{ID: *c.AgendaID}
	if ok, err := agenda.Load(repo); err != nil {
		return err
	} else if !ok {
		return errors.New(pkg.ErrAgendaNotFound)
	}
	return nil
}

// formatValue is a method that formats the value of the contract
func (c *InvoiceItem) formatValue(filled bool) error {
	if math.IsNaN(c.Value) {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrInvalidValue)
	}
	return nil
}

// formatDescription is a method that formats the description of the contract
func (c *InvoiceItem) formatDescription(filled bool) error {
	description := c.formatString(c.Description)
	if description == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyDescription)
	}
	return nil
}

// validateDuplicity is a method that validates the duplicity of a client
func (c *InvoiceItem) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&InvoiceItem{}, c.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}

// formatString is a method that formats a string
func (c *InvoiceItem) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}
