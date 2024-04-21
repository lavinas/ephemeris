package domain

import (
	"strconv"
	"time"
	"errors"
	"fmt"
	"strings"
	"regexp"
	"slices"

	"github.com/lavinas/ephemeris/pkg"
	"github.com/lavinas/ephemeris/internal/port"
)

var (
	status = []string{pkg.InvoiceStatusActive, pkg.InvoiceStatusCanceled}
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
	ID            string    `gorm:"type:varchar(25); primaryKey"`
	Date          time.Time `gorm:"type:date; not null"`
	ClientID      string    `gorm:"type:varchar(25); not null"`
	Value         float64   `gorm:"type:numeric(20,2); not null"`
	Status        string    `gorm:"type:varchar(25), not null"`
	SendStatus    string    `gorm:"type:varchar(25), not null"`
	PaymentStatus string    `gorm:"type:varchar(25), not null"`
}

// NewInvoice creates a new invoice domain entity
func NewInvoice(id, clientID, date, value, status, sendstatus, paymentstatus string) *Invoice {
	invoice := &Invoice{}
	invoice.ID = id
	invoice.ClientID = clientID
	local, _ := time.LoadLocation(pkg.Location)
	invoice.Date, _ = time.ParseInLocation(pkg.DateFormat, date, local)
	invoice.Value, _ = strconv.ParseFloat(value, 64)
	invoice.Status = status
	invoice.SendStatus = sendstatus
	invoice.PaymentStatus = paymentstatus
	return invoice
}

// Format formats the invoice
func (i *Invoice) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	// noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := i.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := i.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	return nil
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
	if len(id) > 25 {
		return errors.New(pkg.ErrLongID)
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

func (c *Invoice) validateClientID() error {
	if !slices.Contains(status, c.Status) {
		return errors.New(pkg.ErrInvalidStatus)
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