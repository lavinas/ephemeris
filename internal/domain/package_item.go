package domain

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// TODO: create command join to add a item to a package
// PackageItem represents the package item entity
type PackageItem struct {
	ID        string   `gorm:"type:varchar(100); primaryKey"`
	PackageID string   `gorm:"type:varchar(50); not null; index"`
	ServiceID string   `gorm:"type:varchar(50); not null; index"`
	Price     *float64 `gorm:"type:decimal(10,2); not null"`
}

// NewPackageItem creates a new package item
func NewPackageItem(id, packageID, serviceID, price string) *PackageItem {
	var p *float64
	if r, err := strconv.ParseFloat(price, 64); err == nil {
		p = &r
	}
	return &PackageItem{
		ID:        id,
		PackageID: packageID,
		ServiceID: serviceID,
		Price:     p,
	}
}

// Format is a method that formats the package item entity
func (p *PackageItem) Format(repo port.Repository, args ...string) error {
	msg := ""
	filled := slices.Contains(args, "filled")
	if err := p.formatID(filled); err != nil {
		msg = err.Error()
	}
	if err := p.formatPackageID(repo, filled); err != nil {
		msg += err.Error()
	}
	if err := p.formatServiceID(repo, filled); err != nil {
		msg += err.Error()
	}
	if err := p.formatPrice(filled); err != nil {
		msg += err.Error()
	}
	if msg != "" {
		return errors.New(msg)
	}
	return nil
}

// Exists is a method that checks if the contract exists
func (p *PackageItem) Load(repo port.Repository) (bool, error) {
	return repo.Get(p, p.ID)
}

// GetID is a method that returns the id of the contract
func (p *PackageItem) GetID() string {
	return p.ID
}

// Get is a method that returns the contract
func (p *PackageItem) Get() port.Domain {
	return p
}

// GetEmpty is a method that returns an empty contract
func (p *PackageItem) GetEmpty() port.Domain {
	return &PackageItem{}
}

// TableName is a method that returns the table name of the contract
func (p *PackageItem) TableName() string {
	return "package_item"
}

// FormatID is a method that formats the package item entity
func (p *PackageItem) formatID(filled bool) error {
	id := p.formatString(p.ID)

	if id == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	if len(id) > 100 {
		return errors.New(pkg.ErrLongID)
	}
	if len(strings.Split(id, " ")) > 1 {
		return errors.New(pkg.ErrInvalidID)
	}
	p.ID = strings.ToLower(id)
	return nil
}

// FormatPackageID is a method that formats the package item entity
func (p *PackageItem) formatPackageID(repo port.Repository, filled bool) error {
	if p.PackageID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyPackageID)
	}
	pack := &Package{ID: p.PackageID}
	if exists, err := pack.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrPackageNotFound)
	}
	return nil
}

// FormatServiceID is a method that formats the package item entity
func (p *PackageItem) formatServiceID(repo port.Repository, filled bool) error {
	serviceID := p.formatString(p.ServiceID)
	if serviceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrServiceIDNotProvided)
	}
	service := &Service{ID: p.ServiceID}
	service.Format(repo, "filled")
	if exists, err := service.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrServiceNotFound)
	}
	return nil
}

// FormatPrice is a method that formats the package item entity
func (p *PackageItem) formatPrice(filled bool) error {
	if p.Price == nil {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyUnitPrice)
	}
	if *p.Price < 0 {
		return errors.New(pkg.ErrInvalidUnitPrice)
	}
	return nil
}

// formatString is a method that formats a string
func (c *PackageItem) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}
