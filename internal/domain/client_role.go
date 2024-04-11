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

var (
	// Roles is a slice that contains the roles for a client
	Roles       = []string{pkg.RoleClient, pkg.RoleLiable, pkg.RolePayer}
	DefaultRole = "client"
)

// ClientRole is a struct that represents the roles of a client
type ClientRole struct {
	ID       string    `gorm:"type:varchar(100); primaryKey"`
	Date     time.Time `gorm:"type:datetime; not null"`
	ClientID string    `gorm:"type:varchar(25); not null; index"`
	Role     string    `gorm:"type:varchar(25); not null; index"`
	RefID    string    `gorm:"type:varchar(25); null; index"`
	Ref      *Client   `gorm:"foreignKey:RefID;associationForeignKey:ID"`
}

// NewClientRole is a function that creates a new client role
func NewClientRole(ID string, date string, clientID string, role string, refID string) *ClientRole {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	return &ClientRole{
		ID:       ID,
		Date:     fdate,
		ClientID: clientID,
		Role:     role,
		RefID:    refID,
	}
}

// Format is a method that formats the client role
func (c *ClientRole) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	msg := ""
	if err := c.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatClientID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatRole(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatRefID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := c.validateDuplicity(repo, slices.Contains(args, "noduplicity")); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if the client role exists
func (c *ClientRole) Exists(repo port.Repository) (bool, error) {
	return repo.Get(c, c.ID)
}

// GetID is a method that returns the id of the client
func (c *ClientRole) GetID() string {
	return c.ID
}

// Get is a method that returns the client
func (c *ClientRole) Get() port.Domain {
	return c
}

// GetEmpty is a method that returns an empty client with just id
func (c *ClientRole) GetEmpty() port.Domain {
	return &ClientRole{}
}

// TableName returns the table name for database
func (b *ClientRole) TableName() string {
	return "client_role"
}

// formatString is a method that formats a string
func (c *ClientRole) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}

// formatID is a method that formats the id field
func (c *ClientRole) formatID(filled bool) error {
	id := c.formatString(c.ID)
	if id == "" {
		if filled {
			return nil
		}
		c.ID = c.mountID(c.ClientID, c.Role, c.RefID)
		return nil
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

// mountID is a method that mounts the id field
func (c *ClientRole) mountID(client_id string, role string, ref_id string) string {
	return strings.ToLower(client_id + "-" + role + "-" + ref_id)
}

// FormatDate is a method that formats the date field
func (c *ClientRole) formatDate(filled bool) error {
	date := c.Date
	if date.IsZero() {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrInvalidDateFormat)
	}
	c.Date = date
	return nil
}

// formatClientID is a method that formats the client id field
func (c *ClientRole) formatClientID(filled bool) error {
	clientID := c.formatString(c.ClientID)
	if clientID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrClientIDNotProvided)
	}
	if len(clientID) > 25 {
		return errors.New(pkg.ErrLongClientID)
	}
	if len(strings.Split(clientID, " ")) > 1 {
		return errors.New(pkg.ErrInvalidClientID)
	}
	c.ClientID = strings.ToLower(clientID)
	return nil
}

// formatRole is a method that formats the role field
func (c *ClientRole) formatRole(filled bool) error {
	role := c.formatString(c.Role)
	if role == "" {
		if filled {
			return nil
		}
		c.Role = DefaultRole
		return nil
	}
	if !slices.Contains(Roles, role) {
		return errors.New(pkg.ErrInvalidRole)
	}
	c.Role = strings.ToLower(role)
	return nil
}

// formatRefID is a method that formats the ref id field
func (c *ClientRole) formatRefID(repo port.Repository, filled bool) error {
	if c.Ref != nil {
		c.RefID = c.Ref.ID
	}
	refID := c.formatString(c.RefID)
	if refID == "" {
		if filled {
			return nil
		}
		if c.Role == pkg.RoleClient {
			c.RefID = c.ClientID
			return nil
		}
		return errors.New(pkg.ErrRefIDNotProvided)
	}
	if len(refID) > 25 {
		return errors.New(pkg.ErrLongRefID)
	}
	if len(strings.Split(refID, " ")) > 1 {
		return errors.New(pkg.ErrInvalidRefID)
	}
	c.RefID = strings.ToLower(refID)
	if c.Role == pkg.RoleClient {
		return nil
	}
	clientRoleId := c.mountID(c.RefID, pkg.RoleClient, c.RefID)
	if b, err := repo.Get(&ClientRole{}, clientRoleId); err != nil {
		return err
	} else if !b {
		return errors.New(pkg.ErrRefNotFound)
	}
	return nil
}

// validateDuplicity is a method that validates the duplicity of the client role
func (c *ClientRole) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity || c.Role == pkg.RoleClient {
		return nil
	}
	ok, err := repo.Get(&ClientRole{}, c.ID)
	if err != nil {
		return err
	}
	if ok {
		return errors.New(pkg.ErrDuplicatedRole)
	}
	return nil
}
