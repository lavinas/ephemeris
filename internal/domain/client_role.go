package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
)

var (
	// Roles is a slice that contains the roles for a client
	Roles = []string{port.RoleClient, port.RoleResponsable, port.RolePayer}
)

// ClientRole is a struct that represents the roles of a client
type ClientRole struct {
	ID       string    `gorm:"type:varchar(100); primaryKey"`
	Date     time.Time `gorm:"type:datetime; not null"`
	ClientID string    `gorm:"type:varchar(25); not null; index"`
	Client   *Client   `gorm:"foreignKey:ClientID;associationForeignKey:ID"`
	Role     string    `gorm:"type:varchar(25); not null; index"`
	RefID    string    `gorm:"type:varchar(25); null; index"`
	Ref      *Client   `gorm:"foreignKey:RefID;associationForeignKey:ID"`
}

// NewClientRole is a function that creates a new client role
func NewClientRole(ID string, date string, clientID string, role string, refID string) *ClientRole {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(port.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(port.DateFormat, date, local); err != nil {
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
func (c *ClientRole) Format(args ...string) error {
	if c.Client != nil {
		c.ClientID = c.Client.ID
	}
	if c.Ref != nil {
		c.RefID = c.Ref.ID
	}
	filled := slices.Contains(args, "filled")
	c.ID = c.formatString(c.ID)
	c.ClientID = c.formatString(c.ClientID)
	c.Role = c.formatString(c.Role)
	c.RefID = c.formatString(c.RefID)
	if c.ID == "" && !filled {
		return errors.New(port.ErrIdUninformed)
	}
	if c.Date == (time.Time{}) && !filled {
		return errors.New(port.ErrInvalidDateFormat)
	}
	if c.ClientID == "" && !filled {
		return errors.New(port.ErrClientIDNotProvided)
	}
	if c.Role == "" && !filled {
		return errors.New(port.ErrRoleNotProvided)
	}
	if c.Role != "" && !slices.Contains(Roles, c.Role) {
		return errors.New(port.ErrInvalidRole)
	}
	if c.RefID == "" && !filled {
		return errors.New(port.ErrRefIDNotProvided)
	}
	if c.Role != "" && c.Role != "client" && c.RefID == c.ClientID {
		return errors.New(port.ErrInvalidReference)
	}
	return nil
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
