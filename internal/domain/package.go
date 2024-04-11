package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Package represents the package entity
type Package struct {
	ID           string    `gorm:"type:varchar(25); primaryKey"`
	Date         time.Time `gorm:"type:datetime; not null; index"`
	Name         string    `gorm:"type:varchar(100); not null; index"`
	ServiceID    string    `gorm:"type:varchar(25); not null; index"`
	RecurrenceID string    `gorm:"type:varchar(25); not null; index"`
	PriceID      string    `gorm:"type:varchar(25); not null; index"`
	BillingType  string    `gorm:"type:varchar(20); not null; index"`
}

var (
	// BillingTypes is a map that contains all billing types
	BillingTypes = map[string]string{
		// pre-paid represents that client paid before the service
		"pre-paid": "pre-paid",
		// pos-paid represents that client paid after the service
		"pos-paid": "pos-paid",
		// pos-session represents that client paid after the service if the session is done
		"pos-session": "pos-session",
	}
)

// NewPackage creates a new package
func NewPackage(id string, date string, name string, serviceID string, recurrenceID string, priceID string, billingType string) *Package {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	return &Package{
		ID:           id,
		Date:         fdate,
		Name:         name,
		ServiceID:    serviceID,
		RecurrenceID: recurrenceID,
		PriceID:      priceID,
		BillingType:  billingType,
	}
}

// Validate is a method that validates the package entity
func (p *Package) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	msg := ""
	if err := p.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatServiceID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatRecurrenceID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatPriceID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatBillingType(filled); err != nil {
		msg += err.Error() + " | "
	}
	return nil
}

// Exists is a method that checks if the contract exists
func (p *Package) Exists(repo port.Repository) (bool, error) {
	return repo.Get(&Package{}, p.ID)
}

// GetID is a method that returns the id of the contract
func (p *Package) GetID() string {
	return p.ID
}

// Get is a method that returns the contract
func (p *Package) Get() port.Domain {
	return p
}

// GetEmpty is a method that returns an empty contract
func (p *Package) GetEmpty() port.Domain {
	return &Package{}
}

// TableName is a method that returns the table name of the contract
func (p *Package) TableName() string {
	return "package"
}

// formatID is a method that formats the id of the package
func (p *Package) formatID(filled bool) error {
	id := p.formatString(p.ID)
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
	p.ID = strings.ToLower(id)
	return nil
}

// formatDate is a method that formats the date of the package
func (p *Package) formatDate(filled bool) error {
	if filled && p.Date.IsZero() {
		return nil
	}
	if p.Date.IsZero() {
		return errors.New(pkg.ErrInvalidDateFormat)
	}
	return nil
}

// formatServiceID is a method that formats the service id of the package
func (p *Package) formatServiceID(repo port.Repository, filled bool) error {
	serviceID := p.formatString(p.ServiceID)
	if serviceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrServiceIDNotProvided)
	}
	service := &Service{ID: p.ServiceID}
	service.Format(repo, "filled")
	if exists, err := service.Exists(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrServiceNotFound)
	}
	return nil
}

// formatRecurrenceID is a method that formats the recurrence id of the contract
func (p *Package) formatRecurrenceID(repo port.Repository, filled bool) error {
	recurrenceID := p.formatString(p.RecurrenceID)
	if recurrenceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrRecurrenceIDNotProvided)
	}
	recurrence := &Recurrence{ID: p.RecurrenceID}
	recurrence.Format(repo, "filled")
	if exists, err := recurrence.Exists(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrRecurrenceNotFound)
	}
	return nil
}

// formatPriceID is a method that formats the price id of the contract
func (p *Package) formatPriceID(repo port.Repository, filled bool) error {
	priceID := p.formatString(p.PriceID)
	if priceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrPriceIDNotProvided)
	}
	price := &Price{ID: p.PriceID}
	price.Format(repo, "filled")
	if exists, err := price.Exists(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrPriceNotFound)
	}
	return nil
}

// formatBillingType is a method that formats the billing type of the contract
func (p *Package) formatBillingType(filled bool) error {
	p.BillingType = p.formatString(p.BillingType)
	if p.BillingType == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyBillingType)
	}
	p.BillingType = strings.ToLower(p.BillingType)
	if _, ok := BillingTypes[p.BillingType]; !ok {
		bt := ""
		for k := range BillingTypes {
			bt += k + ", "
		}
		return fmt.Errorf(pkg.ErrInvalidBillingType, bt[:len(bt)-2])
	}
	return nil
}

// formatString is a method that formats a string
func (c *Package) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}
