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

var (
	// BillingTypes is a map that contains all billing types
	billingTypes = []string{
		"pre-paid", "pos-paid", "pos-session", "per-session",
	}
)

// Contract represents the contract entity
type Contract struct {
	ID          string     `gorm:"type:varchar(25); primaryKey"`
	Date        time.Time  `gorm:"type:datetime; not null; index"`
	ClientID    string     `gorm:"type:varchar(25); not null; index"`
	SponsorID   *string    `gorm:"type:varchar(25); null; index"`
	PackageID   string     `gorm:"type:varchar(25); not null; index"`
	BillingType string     `gorm:"type:varchar(25); not null; index"`
	DueDay      *int64     `gorm:"type:numeric(20); null; index"`
	Start       time.Time  `gorm:"type:datetime; not null; index"`
	End         *time.Time `gorm:"type:datetime; null; index"`
	Bond        *string    `gorm:"type:varchar(25); null; index"`
	Locked      *bool      `gorm:"type:boolean;null; index"`
}

// NewContract creates a new contract
func NewContract(id, date, clientID, SponsorID, packageID, billingType, dueDay, start, end, bond string) *Contract {
	contract := &Contract{}
	contract.ID = id
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	contract.Date, _ = time.ParseInLocation(pkg.DateFormat, date, local)
	contract.ClientID = clientID
	contract.PackageID = packageID
	contract.BillingType = billingType
	contract.Start, _ = time.ParseInLocation(pkg.DateTimeFormat, start, local)
	if SponsorID != "" {
		contract.SponsorID = &SponsorID
	}
	if d, err := strconv.ParseInt(dueDay, 10, 64); err == nil {
		contract.DueDay = &d
	}
	if d, err := time.ParseInLocation(pkg.DateFormat, end, local); err == nil {
		contract.End = &d
	}
	if bond != "" {
		contract.Bond = &bond
	}
	return contract
}

// Format is a method that formats the contract
func (c *Contract) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := c.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatClientID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatSponsorID(repo); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatPackageID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatBillingType(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatDueDay(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatStart(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatEnd(); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatBond(repo); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.validateDuplicity(repo, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if the contract exists
func (c *Contract) Load(repo port.Repository) (bool, error) {
	return repo.Get(c, c.ID)
}

// GetID is a method that returns the id of the contract
func (c *Contract) GetID() string {
	return c.ID
}

// Get is a method that returns the contract
func (c *Contract) Get() port.Domain {
	return c
}

// GetEmpty is a method that returns an empty contract
func (c *Contract) GetEmpty() port.Domain {
	return &Contract{}
}

// Lock is a method that locks the contract
func (c *Contract) Lock(repo port.Repository) error {
	var locked = true
	c.Locked = &locked
	if err := repo.Begin(); err != nil {
		return err
	}
	defer repo.Rollback()
	if err := repo.Save(c); err != nil {
		return err
	}
	if err := repo.Commit(); err != nil {
		return err
	}
	return nil
}

// IsLocked is a method that checks if the contract is locked
func (c *Contract) IsLocked() bool {
	return c.Locked != nil && *c.Locked
}

// Unlock is a method that unlocks the contract
func (c *Contract) Unlock(repo port.Repository) error {
	c.Locked = nil
	if err := repo.Begin(); err != nil {
		return err
	}
	if err := repo.Save(c); err != nil {
		return err
	}
	if err := repo.Commit(); err != nil {
		return err
	}
	return nil
}

// TableName is a method that returns the table name of the contract
func (c *Contract) TableName() string {
	return "contract"
}

// formatID is a method that formats the id of the contract
func (c *Contract) formatID(filled bool) error {
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
func (c *Contract) formatDate(filled bool) error {
	if filled && c.Date.IsZero() {
		return nil
	}
	if c.Date.IsZero() {
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
	}
	return nil
}

// formatClientID is a method that formats the client id of the contract
func (c *Contract) formatClientID(repo port.Repository, filled bool) error {
	clientID := c.formatString(c.ClientID)
	if clientID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrClientIDNotProvided)
	}
	client := &Client{ID: c.ClientID}
	client.Format(repo, "filled")
	if exists, err := client.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrClientNotFound)
	}
	return nil
}

// formatSponsorID is a method that formats the sponsor id of the contract
func (c *Contract) formatSponsorID(repo port.Repository) error {
	if c.SponsorID == nil {
		return nil
	}
	client := &Client{ID: c.formatString(*c.SponsorID)}
	client.Format(repo, "filled")
	if exists, err := client.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrSponsorNotFound)
	}
	return nil
}

// formatServiceID is a method that formats the service id of the contract
func (c *Contract) formatPackageID(repo port.Repository, filled bool) error {
	packageID := c.formatString(c.PackageID)
	if packageID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrPackageIDNotProvided)
	}
	pack := &Package{ID: c.PackageID}
	pack.Format(repo, "filled")
	if exists, err := pack.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrPackageNotFound)
	}
	return nil
}

// formatBillingType is a method that formats the billing type of the contract
func (c *Contract) formatBillingType(filled bool) error {
	c.BillingType = c.formatString(c.BillingType)
	if c.BillingType == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyBillingType)
	}
	c.BillingType = strings.ToLower(c.BillingType)
	if !slices.Contains(billingTypes, c.BillingType) {
		return fmt.Errorf(pkg.ErrInvalidBillingType, strings.Join(billingTypes, ", "))
	}
	return nil
}

// formatDueDay is a method that formats the due day of the contract
func (c *Contract) formatDueDay(filled bool) error {
	if c.DueDay == nil {
		if filled {
			return nil
		}
		if c.BillingType == pkg.BillingTypePerSession {
			return nil
		}
		return errors.New(pkg.ErrDueDayNotProvided)
	}
	if *c.DueDay <= 0 || *c.DueDay > 31 {
		return errors.New(pkg.ErrInvalidDueDay)
	}
	return nil
}

// formatStart is a method that formats the start of the contract
func (c *Contract) formatStart(filled bool) error {
	if c.Start.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidStartDate, pkg.DateTimeFormat)
	}
	return nil
}

// formatEnd is a method that formats the end of the contract
func (c *Contract) formatEnd() error {
	if c.End == nil {
		return nil
	}
	if c.End.IsZero() {
		return fmt.Errorf(pkg.ErrInvalidEndDate, pkg.DateFormat)
	}
	return nil
}

// formatBond is a method that formats the bond of the contract
func (c *Contract) formatBond(repo port.Repository) error {
	if c.Bond == nil {
		return nil
	}
	linkContract := &Contract{ID: *c.Bond}
	linkContract.Format(repo, "filled")
	if exists, err := linkContract.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrBondNotFound)
	}
	return nil
}

// formatString is a method that formats a string
func (c *Contract) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}

// validateDuplicity is a method that validates the duplicity of a client
func (c *Contract) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Contract{}, c.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}
