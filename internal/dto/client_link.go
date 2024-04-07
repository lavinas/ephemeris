package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

// ClientLinkIn represents the dto for linking a client
type ClientLinkIn struct {
	Object string `json:"-" command:"name:client;key"`
	Action string `json:"-" command:"name:link;key"`
	ID     string `json:"referrer" command:"name:id"`
	Ref    string `json:"ref_id" command:"name:ref"`
	Role   string `json:"role" command:"name:role"`
	Date   string `json:"date" command:"name:date"`
}

// ClientLinkOut represents the dto for linking a client
type ClientLinkOut struct {
	Actor string `json:"referrer" command:"name:id"`
	Ref   string `json:"ref_id" command:"name:ref"`
	Role  string `json:"role" command:"name:role"`
	Date  string `json:"date" command:"name:date"`
}

// Validate is a method that validates the dto
func (c *ClientLinkIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (c *ClientLinkIn) GetDomain() []port.Domain {
	if c.Date == "" {
		time.Local, _ = time.LoadLocation(port.Location)
		c.Date = time.Now().Format(port.DateFormat)
	}
	return []port.Domain{
		domain.NewClientRole("", c.Date, c.ID, c.Role, c.Ref),
	}
}

// GetOut is a method that returns the dto out
func (c *ClientLinkIn) GetOut() port.DTOOut {
	return &ClientLinkOut{}
}

// GetDTO is a method that returns the dto
func (c *ClientLinkOut) GetDTO(domainIn interface{}) interface{} {
	slices := domainIn.([]interface{})
	clientRole, ok := slices[0].(*domain.ClientRole)
	if !ok {
		return nil
	}
	return &ClientLinkOut{
		Actor: clientRole.ClientID,
		Ref:   clientRole.RefID,
		Role:  clientRole.Role,
		Date:  clientRole.Date.Format(port.DateFormat),
	}
}

// isEmpty is a method that checks if the dto is empty
func (c *ClientLinkIn) isEmpty() bool {
	return c.ID == "" && c.Ref == "" && c.Role == "" && c.Date == ""
}
