package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type ClientUp struct {
	Object      string `json:"-" command:"name:client;key"`
	Action      string `json:"-" command:"name:up;key"`
	ID          string `json:"id" command:"name:id"`
	Name        string `json:"name" command:"name:name"`
	Responsible string `json:"responsible" command:"name:responsible"`
	Email       string `json:"email" command:"name:email"`
	Phone       string `json:"phone" command:"name:phone"`
	Contact     string `json:"contact" command:"name:contact"`
	Document    string `json:"document" command:"name:document"`
}

// GetDomain is a method that returns a domain representation of the client dto
func (c *ClientUp) GetDomain() port.Domain {
	return domain.NewClient(c.ID, c.Name, c.Responsible, c.Email, c.Phone, c.Contact, c.Document)
}

// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientUp) GetDto(in interface{}) (interface{}, string) {
	d := in.(*domain.Client)
	ret := &ClientUp{
		ID:          d.ID,
		Name:        d.Name,
		Responsible: d.Responsible,
		Email:       d.Email,
		Phone:       d.Phone,
		Contact:     d.Contact,
		Document:    d.Document,
	}
	return ret, pkg.NewCommands().Marshal(ret, "nokeys")
}

// Validate is a method that validates the dto
func (c *ClientUp) Validate() error {
	if c.IsEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	if c.ID == "" {
		return errors.New(port.ErrIdUninformed)
	}
	id := c.ID
	c.ID = ""
	if c.IsEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	c.ID = id
	return nil
}

// IsEmpty is a method that returns true if the dto is empty
func (c *ClientUp) IsEmpty() bool {
	if c.Object == "" && c.Action == "" && c.ID == "" && c.Name == "" && c.Responsible == "" &&
		c.Email == "" && c.Phone == "" && c.Contact == "" && c.Document == "" {
		return true
	}
	return false
}
