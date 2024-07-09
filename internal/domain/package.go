package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Package represents the package entity
type Package struct {
	ID           string    `gorm:"type:varchar(50); primaryKey"`
	Date         time.Time `gorm:"type:datetime; not null; index"`
	RecurrenceID string    `gorm:"type:varchar(50); not null; index"`
	Price        *float64  `gorm:"type:decimal(10,2); index"`
}

// NewPackage creates a new package
func NewPackage(id, date, recurrenceID, packValue string) *Package {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	var err error
	if date != "" {
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	var p *float64
	if r, err := strconv.ParseFloat(packValue, 64); err == nil {
		p = &r
	}
	return &Package{
		ID:           id,
		Date:         fdate,
		RecurrenceID: recurrenceID,
		Price:        p,
	}
}

// Validate is a method that validates the package entity
func (p *Package) Format(repo port.Repository, tx string, args ...string) error {
	filled := slices.Contains(args, "filled")
	msg := ""
	if err := p.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatRecurrenceID(repo, tx, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatPrice(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.validateDuplicity(repo, tx, slices.Contains(args, "noduplicity")); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if the contract exists
func (p *Package) Load(repo port.Repository, tx string) (bool, error) {
	return repo.Get(p, p.ID, tx)
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

// GetService is a method that returns the service of the package
func (p *Package) GetServices(repo port.Repository, tx string) ([]*Service, []*float64, error) {
	services := []*Service{}
	prices := []*float64{}
	i, _, err := repo.Find(&PackageItem{PackageID: p.ID}, -1, tx)
	if err != nil {
		return nil, nil, err
	}
	items := i.(*[]PackageItem)
	if len(*items) == 0 {
		return nil, nil, errors.New(pkg.ErrServiceNotFound)
	}
	for _, item := range *items {
		service, error := item.GetService(repo, tx)
		if error != nil {
			return nil, nil, error
		}
		services = append(services, service)
		if p.Price != nil && *p.Price > 0 {
			prices = append(prices, nil)
		} else {
			prices = append(prices, item.Price)
		}
	}
	return services, prices, nil
}

// GetRecurrence is a method that returns the recurrence of the package
func (p *Package) GetRecurrence(repo port.Repository, tx string) (*Recurrence, error) {
	if p.RecurrenceID == "" {
		if ok, err := p.Load(repo, tx); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New(pkg.ErrPackageNotFound)
		}
	}
	recurrence := &Recurrence{ID: p.RecurrenceID}
	if ok, err := recurrence.Load(repo, tx); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New(pkg.ErrRecurrenceNotFound)
	}
	return recurrence, nil
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
	if len(id) > 50 {
		return errors.New(pkg.ErrLongID50)
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
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
	}
	return nil
}

// formatRecurrenceID is a method that formats the recurrence id of the contract
func (p *Package) formatRecurrenceID(repo port.Repository, tx string, filled bool) error {
	recurrenceID := p.formatString(p.RecurrenceID)
	if recurrenceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrRecurrenceIDNotProvided)
	}
	recurrence := &Recurrence{ID: p.RecurrenceID}
	recurrence.Format(repo, tx, "filled")
	if exists, err := recurrence.Load(repo, tx); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrRecurrenceNotFound)
	}
	return nil
}

// formatPriceID is a method that formats the price id of the contract
func (p *Package) formatPrice(filled bool) error {
	if p.Price == nil {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrPriceIDNotProvided)
	}
	if *p.Price < 0 {
		return errors.New(pkg.ErrInvalidPackPrice)
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

// validateDuplicity is a method that validates the duplicity of a client
func (c *Package) validateDuplicity(repo port.Repository, tx string, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Package{}, c.ID, tx)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}
