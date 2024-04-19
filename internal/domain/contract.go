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
	BillingTypes = map[string]string{
		// pre-paid represents that client paid before the service
		"pre-paid": "pre-paid",
		// pos-paid represents that client paid after the service
		"pos-paid": "pos-paid",
		// pos-session represents that client paid after the service if the session is done
		"pos-session": "pos-session",
		// per-session represents that client paid for each session
		"per-session": "per-session",
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
	DueDay      *int64     `gorm:"type:numeric(20), null; index"`
	Start       time.Time  `gorm:"type:datetime; not null; index"`
	End         *time.Time `gorm:"type:datetime; null; index"`
	Bond        *string    `gorm:"type:varchar(25); null; index"`
}

// NewContract creates a new contract
func NewContract(id string, date string, clientID string, SponsorID string, packageID string,
	billingType string, dueDay string, start string, end string, bond string) *Contract {
	var err error
	contract := &Contract{}
	contract.ID = id
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	contract.Date, err = time.ParseInLocation(pkg.DateFormat, date, local)
	if err != nil {
		contract.Date = time.Time{}
	}
	contract.ClientID = clientID
	if SponsorID != "" {
		contract.SponsorID = &SponsorID
	}
	contract.PackageID = packageID
	contract.BillingType = billingType
	if d, err := strconv.ParseInt(dueDay, 10, 64); err == nil && d >= 0 {
		contract.DueDay = &d
	}
	contract.Start, err = time.ParseInLocation(pkg.DateFormat, start, local)
	if err != nil {
		contract.Start = time.Time{}
	}
	if end != "" {
		*contract.End, err = time.ParseInLocation(pkg.DateFormat, end, local)
		if err != nil {
			contract.End = nil
		}
	}
	if bond != "" {
		contract.Bond = &bond
	}
	return contract
}

// Format is a method that formats the contract
func (c *Contract) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	// noduplicity := slices.Contains(args, "noduplicity")
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
	if err := c.validateDuplicity(repo, false); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if the contract exists
func (c *Contract) Exists(repo port.Repository) (bool, error) {
	return repo.Get(&Contract{}, c.ID)
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
		return errors.New(pkg.ErrInvalidDateFormat)
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
	if exists, err := client.Exists(repo); err != nil {
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
	if exists, err := client.Exists(repo); err != nil {
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
		return errors.New(pkg.ErrServiceIDNotProvided)
	}
	pack := &Package{ID: c.PackageID}
	pack.Format(repo, "filled")
	if exists, err := pack.Exists(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrServiceNotFound)
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
	if _, ok := BillingTypes[c.BillingType]; !ok {
		bt := ""
		for k := range BillingTypes {
			bt += k + ", "
		}
		return fmt.Errorf(pkg.ErrInvalidBillingType, bt[:len(bt)-2])
	}
	return nil
}

// formatDueDay is a method that formats the due day of the contract
func (c *Contract) formatDueDay(filled bool) error {
	if c.DueDay == nil {
		if filled {
			return nil
		}
		if c.BillingType == BillingTypes["per-session"] {
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
		return fmt.Errorf(pkg.ErrInvalidStartDate, pkg.DateFormat)
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
	if exists, err := linkContract.Exists(repo); err != nil {
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
