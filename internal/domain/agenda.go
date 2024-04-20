package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Agenda represents the agenda entity
type Agenda struct {
	ID           string     `gorm:"type:varchar(25); primaryKey"`
	Date         time.Time  `gorm:"type:datetime; not null"`
	ContractID   string     `gorm:"type:varchar(25); not null; index"`
	Start        time.Time  `gorm:"type:datetime; not null"`
	End          *time.Time  `gorm:"type:datetime; not null"`
	Kind         string     `gorm:"type:varchar(25); not null; index"`
	Status       string     `gorm:"type:varchar(25); not null; index"`
	Bond         *Agenda    `gorm:"foreignKey:ID"`
	BillingMonth *time.Time `gorm:"type:numeric(20)"`
}

// NewAgenda creates a new agenda domain entity
func NewAgenda(id, date, contractID, start, end, kind, status, bond, billing string) *Agenda {
	agenda := &Agenda{}
	agenda.ID = id
	local, _ := time.LoadLocation(pkg.Location)
	agenda.Date, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(date), local)
	agenda.ContractID = contractID
	agenda.Start, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(start), local)
	e, err := time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(end), local)
	if err == nil && !e.IsZero() {
		agenda.End = &e
	}
	agenda.Kind = kind
	agenda.Status = status
	if bond != "" {
		agenda.Bond = &Agenda{ID: bond}
	}
	mont, err := time.ParseInLocation(pkg.MonthFormat, billing, local)
	if err == nil &&  !mont.IsZero(){
		agenda.BillingMonth = &mont
	} else {
		mont, err = time.ParseInLocation(pkg.DateFormat, billing, local)
		if err == nil &&  !mont.IsZero(){
			agenda.BillingMonth = &mont
		}
	}
	return agenda
}

// Format formats the agenda
func (a *Agenda) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	// noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := a.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatContractID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatStart(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatEnd(); err != nil {
		msg += err.Error() + " | "
	}
	if msg == "" {
		return nil
	}
	return errors.New(msg)
}

// formatID is a method that formats the id of the contract
func (c *Agenda) formatID(filled bool) error {
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
func (c *Agenda) formatDate(filled bool) error {
	if filled {
		return nil
	}
	if c.Date.IsZero() {
		return errors.New(pkg.ErrInvalidDateFormat)
	}
	return nil
}

// formatContractID is a method that formats the contract id
func (c *Agenda) formatContractID(repo port.Repository, filled bool) error {
	contractID := c.formatString(c.ContractID)
	if contractID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyContractID)
	}
	contract := &Contract{ID: contractID}
	if exists, err := contract.Exists(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrContractNotFound)
	}
	return nil
}

// formatStart is a method that formats the start date of the agenda
func (c *Agenda) formatStart(filled bool) error {
	if filled {
		return nil
	}
	if c.Start.IsZero() {
		return errors.New(pkg.ErrInvalidDateFormat)
	}
	return nil
}

// formatEnd is a method that formats the end date of the agenda
func (c *Agenda) formatEnd() error {
	if c.End == nil {
		return nil
	}
	if c.End.IsZero() {
		return errors.New(pkg.ErrInvalidDateFormat)
	}
	if c.Start.Before(*c.End) {
		return errors.New(pkg.ErrInvalidEnd)
	}
	return nil
}


// formatString is a method that formats a string
func (c *Agenda) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}
