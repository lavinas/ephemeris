package domain

import (
	"time"
	"errors"
	"slices"

	"github.com/lavinas/ephemeris/internal/port"
)

var (
	// Roles is a slice that contains the roles for a client
	Roles = []string{"client", "responsable", "payer"}
)

type ClientRole struct {
	ClientID string    `gorm:"type:varchar(25); primaryKey"`
	Client   *Client   `gorm:"foreignKey:ClientID;associationForeignKey:ID"`
	Role     string    `gorm:"type:varchar(25); primaryKey"`
	RefID    string    `gorm:"type:varchar(25); null"`
	Ref      *Client   `gorm:"foreignKey:RefID;associationForeignKey:ID"`
	Date     time.Time `gorm:"type:datetime; not null"`
}

// NewClientRole is a function that creates a new client role
func NewClientRole(clientID string, role string, refID string, date time.Time) *ClientRole {
	return &ClientRole{
		ClientID: clientID,
		Role:     role,
		RefID:    refID,
	}
}

// Format is a method that formats the client role
func (c *ClientRole) Format() error {
	if c.Client != nil {
		c.ClientID = c.Client.ID
	}
	if c.Ref != nil {
		c.RefID = c.Ref.ID
	}
	if c.Date == (time.Time{}) {
		c.Date = time.Now()
	}
	if c.ClientID == "" {
		return errors.New(port.ErrClientIDNotProvided)
	}
	if c.Role == "" {
		return errors.New(port.ErrRoleNotProvided)
	}
	if !slices.Contains(Roles, c.Role) {
		return errors.New(port.ErrInvalidRole)
	}
	if c.RefID == "" {
		return errors.New(port.ErrRefIDNotProvided)
	}
	if c.Role != "client" && c.RefID == c.ClientID {
		return errors.New(port.ErrInvalidReference)
	}
	return nil
}

// TableName returns the table name for database
func (b *ClientRole) TableName() string {
	return "client_role"
}
