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

// Price represents the price entity
type Price struct {
	ID   string    `gorm:"type:varchar(25); primaryKey"`
	Date time.Time `gorm:"type:datetime; not null; index"`
	Name string    `gorm:"type:varchar(100); not null; index"`
	Unit *float64  `gorm:"type:numeric(20,2); null; index"`
	Pack *float64  `gorm:"type:numeric(20,2); null; index"`
}

// NewPrice is a function that creates a new price
func NewPrice(id string, date string, name string, unit string, pack string) *Price {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	var funit *float64 = nil
	if u, err := strconv.ParseFloat(unit, 64); u >= 0 && err == nil {
		funit = &u
	}
	var fpack *float64 = nil
	if p, err := strconv.ParseFloat(pack, 64); p >= 0 && err == nil {
		fpack = &p
	}
	return &Price{
		ID:   id,
		Date: fdate,
		Name: name,
		Unit: funit,
		Pack: fpack,
	}
}

// Format is a method that formats the price
func (p *Price) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := p.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatName(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.formatUnitAndPack(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := p.validateDuplicity(repo, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if the price exists
func (p *Price) Exists(repo port.Repository) (bool, error) {
	return repo.Get(&Price{}, p.ID)
}

// GetID is a method that returns the id of the price
func (p *Price) GetID() string {
	return p.ID
}

// Get is a method that returns the price
func (p *Price) Get() port.Domain {
	return p
}

// GetEmpty is a method that returns an empty price with just id
func (p *Price) GetEmpty() port.Domain {
	return &Price{}
}

// TableName returns the table name for database
func (p *Price) TableName() string {
	return "price"
}

// formatID is a method that formats the price id
func (p *Price) formatID(filled bool) error {
	p.ID = p.formatString(p.ID)
	if p.ID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	return nil
}

// formatDate is a method that formats the price date
func (p *Price) formatDate(filled bool) error {
	if filled {
		return nil
	}
	if p.Date.IsZero() {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrInvalidDateFormat)
	}
	return nil
}

// formatName is a method that formats the price name
func (p *Price) formatName(filled bool) error {
	p.Name = p.formatString(p.Name)
	if p.Name == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyName)
	}
	return nil
}

// formatUnit is a method that formats the price unit
func (p *Price) formatUnitAndPack(filled bool) error {
	if p.Pack == nil && p.Unit == nil {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyUnitAndPack)
	}
	if p.Pack != nil && p.Unit != nil {
		return errors.New(pkg.ErrDuplicityUnitAndPack)
	}
	return nil
}

// formatString is a method that formats a string
func (p *Price) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}

// validateDuplicity is a method that validates the duplicity of a price
func (p *Price) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Price{}, p.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, p.ID)
	}
	return nil
}
