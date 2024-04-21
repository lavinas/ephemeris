package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
	"fmt"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	kindAgenda   = []string{pkg.AgendaKindSlated, pkg.AgendaKindRescheduled, pkg.AgendaKindExtra}
	statusAgenda = []string{pkg.AgendaStatusSlated, pkg.AgendaStatusDone, pkg.AgendaStatusCanceled, pkg.AgendaStatusOverdue}
)

// Agenda represents the agenda entity
type Agenda struct {
	ID           string     `gorm:"type:varchar(25); primaryKey"`
	Date         time.Time  `gorm:"type:datetime; not null"`
	ContractID   string     `gorm:"type:varchar(25); not null; index"`
	Start        time.Time  `gorm:"type:datetime; not null"`
	End          time.Time  `gorm:"type:datetime;not null"`
	Kind         string     `gorm:"type:varchar(25); not null; index"`
	Status       string     `gorm:"type:varchar(25); not null; index"`
	Bond         *string    `gorm:"type:varchar(25);null; index"`
	BillingMonth *time.Time `gorm:"type:numeric(20);null; index"`
}

// NewAgenda creates a new agenda domain entity
func NewAgenda(id, date, contractID, start, end, kind, status, bond, billing string) *Agenda {
	agenda := &Agenda{}
	agenda.ID = id
	local, _ := time.LoadLocation(pkg.Location)
	agenda.Date, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(date), local)
	agenda.ContractID = contractID
	agenda.Start, _ = time.ParseInLocation(pkg.DateTimeFormat, strings.TrimSpace(start), local)
	agenda.End, _ = time.ParseInLocation(pkg.DateTimeFormat, strings.TrimSpace(end), local)
	agenda.Kind = kind
	agenda.Status = status
	if bond != "" {
		agenda.Bond = &bond
	}
	mont, err := time.ParseInLocation(pkg.MonthFormat, billing, local)
	if err == nil && !mont.IsZero() {
		agenda.BillingMonth = &mont
	} else {
		mont, err = time.ParseInLocation(pkg.DateFormat, billing, local)
		if err == nil && !mont.IsZero() {
			agenda.BillingMonth = &mont
		}
	}
	return agenda
}

// Format formats the agenda
func (a *Agenda) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
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
	if err := a.formatEnd(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatKind(); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatStatus(); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatBond(repo); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatBillingMonth(); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.validateDuplicity(repo, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg == "" {
		return nil
	}
	return errors.New(msg[:len(msg)-3])
}

// Exists is a function that checks if a agenda exists
func (a *Agenda) Exists(repo port.Repository) (bool, error) {
	return repo.Get(&Agenda{}, a.ID)
}

// GetID is a method that returns the id of the client
func (a *Agenda) GetID() string {
	return a.ID
}

// Get is a method that returns the client
func (a *Agenda) Get() port.Domain {
	return a
}

// GetEmpty is a method that returns an empty client with just id
func (a *Agenda) GetEmpty() port.Domain {
	return &Agenda{}
}

// TableName returns the table name for database
func (a *Agenda) TableName() string {
	return "agenda"
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
	if c.Date.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
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
		return fmt.Errorf(pkg.ErrInvalidStartDate, pkg.DateTimeFormat)
	}
	return nil
}

// formatEnd is a method that formats the end date of the agenda
func (c *Agenda) formatEnd(filled bool) error {	
	if c.End.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidEndDate, pkg.DateTimeFormat)
	}
	if c.Start.After(c.End) {
		return errors.New(pkg.ErrStartAfterEndDate)
	}
	return nil
}

// formatKind is a method that formats the kind of the agenda
func (c *Agenda) formatKind() error {
	kind := c.formatString(c.Kind)
	if kind == "" {
		return errors.New(pkg.ErrEmptyKind)
	}
	if !slices.Contains(kindAgenda, kind) {
		return errors.New(pkg.ErrInvalidKind)
	}
	c.Kind = kind
	return nil
}

// formatStatus is a method that formats the status of the agenda
func (c *Agenda) formatStatus() error {
	status := c.formatString(c.Status)
	if status == "" {
		return errors.New(pkg.ErrEmptyStatus)
	}
	if !slices.Contains(statusAgenda, status) {
		return errors.New(pkg.ErrInvalidStatus)
	}
	c.Status = status
	return nil
}

// formatBond is a method that formats the bond of the agenda
func (c *Agenda) formatBond(repo port.Repository) error {
	if c.Bond == nil {
		return nil
	}
	bond := &Agenda{ID: *c.Bond}
	if exists, err := bond.Exists(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrBondNotFound)
	}
	return nil
}

// formatBillingMonth is a method that formats the billing month of the agenda
func (c *Agenda) formatBillingMonth() error {
	if c.BillingMonth == nil {
		return nil
	}
	if c.BillingMonth.IsZero() {
		return fmt.Errorf(pkg.ErrInvalidBillingMonth, pkg.MonthFormat)
	}
	return nil
}

// validateDuplicity is a method that validates the duplicity of a client
func (c *Agenda) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Agenda{}, c.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
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
