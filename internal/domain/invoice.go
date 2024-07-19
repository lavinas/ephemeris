package domain

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	invoiceStatus = []string{pkg.InvoiceStatusActive, pkg.InvoiceStatusCanceled}
	paymentstatus = []string{pkg.InvoicePaymentStatusOpen,
		pkg.InvoicePaymentStatusPaid,
		pkg.InvoicePaymentStatusLate,
		pkg.InvoicePaymentStatusRefund,
		pkg.InvoicePaymentStatusOver,
		pkg.InvoicePaymentStatusUnder,
	}
	sendstatus = []string{pkg.InvoiceSendStatusNotSent,
		pkg.InvoiceSendStatusSent,
		pkg.InvoiceSendStatusViewed,
	}
)

// Invoice represents the invoice entity
type Invoice struct {
	ID            string    `gorm:"type:varchar(150); primaryKey"`
	Date          time.Time `gorm:"type:datetime; not null; index"`
	ClientID      string    `gorm:"type:varchar(50); not null; index"`
	Value         float64   `gorm:"type:numeric(20,2); not null; index"`
	Status        string    `gorm:"type:varchar(50); not null; index"`
	SendStatus    string    `gorm:"type:varchar(50); not null; index"`
	PaymentStatus string    `gorm:"type:varchar(50); not null; index"`
}

// NewInvoice creates a new invoice domain entity
func NewInvoice(id, clientID, date, value, status, sendstatus, paymentstatus string) *Invoice {
	invoice := &Invoice{}
	invoice.ID = id
	invoice.ClientID = clientID
	local, _ := time.LoadLocation(pkg.Location)
	invoice.Date, _ = time.ParseInLocation(pkg.DateFormat, date, local)
	var err error
	if invoice.Value, err = strconv.ParseFloat(value, 64); err != nil {
		invoice.Value = math.NaN()
	}
	invoice.Status = status
	invoice.SendStatus = sendstatus
	invoice.PaymentStatus = paymentstatus
	return invoice
}

// Format formats the invoice
func (i *Invoice) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := i.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatClientID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatValue(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatSendStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatPaymentStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	tx := repo.Begin()
	defer repo.Rollback(tx)
	if err := i.validateDuplicity(repo, tx, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a function that checks if a client exists
func (c *Invoice) Load(repo port.Repository) (bool, error) {
	tx := repo.Begin()
	defer repo.Rollback(tx)
	return repo.Get(tx, c, c.ID, false)
}

// GetID is a method that returns the id of the client
func (c *Invoice) GetID() string {
	return c.ID
}

// Get is a method that returns the client
func (c *Invoice) Get() port.Domain {
	return c
}

// GetEmpty is a method that returns an empty client with just id
func (c *Invoice) GetEmpty() port.Domain {
	return &Invoice{}
}

// TableName returns the table name for database
func (b *Invoice) TableName() string {
	return "invoice"
}

// formatID is a method that formats the id of the contract
func (c *Invoice) formatID(filled bool) error {
	id := c.formatString(c.ID)
	if id == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	if len(id) > 150 {
		return errors.New(pkg.ErrLongID150)
	}
	if len(strings.Split(id, " ")) > 1 {
		return errors.New(pkg.ErrInvalidID)
	}
	c.ID = strings.ToLower(id)
	return nil
}

// formatDate is a method that formats the date of the contract
func (c *Invoice) formatDate(filled bool) error {
	if c.Date.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
	}
	return nil
}

// formatClientID is a method that formats the client id of the contract
func (c *Invoice) formatClientID(repo port.Repository, filled bool) error {
	c.ClientID = c.formatString(c.ClientID)
	if c.ClientID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyClientID)
	}
	client := &Client{ID: c.ClientID}
	if exists, err := client.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrClientNotFound)
	}
	return nil
}

// formatValue is a method that formats the value of the contract
func (c *Invoice) formatValue(filled bool) error {
	if math.IsNaN(c.Value) {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrInvalidValue)
	}
	return nil
}

// formatStatus is a method that formats the status of the contract
func (c *Invoice) formatStatus(filled bool) error {
	c.Status = c.formatString(c.Status)
	if c.Status == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyStatus)
	}
	if !slices.Contains(invoiceStatus, c.Status) {
		status := strings.Join(cycles, ", ")
		return fmt.Errorf(pkg.ErrInvalidStatus, status[:len(status)-2])
	}
	return nil
}

// formatSendStatus is a method that formats the send status of the contract
func (c *Invoice) formatSendStatus(filled bool) error {
	c.SendStatus = c.formatString(c.SendStatus)
	if c.SendStatus == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptySendStatus)
	}
	if !slices.Contains(sendstatus, c.SendStatus) {
		return fmt.Errorf(pkg.ErrInvalidSendStatus, strings.Join(cycles, ", "))
	}
	return nil
}

// formatPaymentStatus is a method that formats the payment status of the contract
func (c *Invoice) formatPaymentStatus(filled bool) error {
	c.PaymentStatus = c.formatString(c.PaymentStatus)
	if c.PaymentStatus == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyPaymentStatus)
	}
	if !slices.Contains(paymentstatus, c.PaymentStatus) {
		return fmt.Errorf(pkg.ErrInvalidPaymentStatus, strings.Join(cycles, ", "))
	}
	return nil
}

// formatString is a method that formats a string
func (c *Invoice) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}

// validateDuplicity is a method that validates the duplicity of a client
func (c *Invoice) validateDuplicity(repo port.Repository, tx interface{}, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(tx, &Invoice{}, c.ID, false)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}
