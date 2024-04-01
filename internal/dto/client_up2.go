package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type ClientUp2 struct {
	Object      string `json:"-" command:"name:client2;key"`
	Action      string `json:"-" command:"name:up;key"`
	ID          string `json:"id" command:"name:id"`
	Date        string `json:"contact" command:"name:date"`
	Name        string `json:"name" command:"name:name"`
	Email       string `json:"email" command:"name:email"`
	Phone       string `json:"phone" command:"name:phone"`
	Document    string `json:"document" command:"name:document"`
}

// GetDomain is a method that returns a domain representation of the client dto
func (c *ClientUp2) GetDomain() port.Domain {
	return domain.NewClient2(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, "")
}

// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientUp2) GetDto(in interface{}) (interface{}, string) {
	d := in.(*domain.Client2)
	ret := &ClientUp2{
		ID:          d.ID,
		Date:        d.Date.Format(port.DateFormat),
		Name:        d.Name,
		Email:       d.Email,
		Phone:       d.Phone,
		Document:    d.Document,
	}
	return ret, pkg.NewCommands().Marshal(ret, "nokeys")
}

// Validate is a method that validates the dto
func (c *ClientUp2) Validate() error {
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
func (c *ClientUp2) IsEmpty() bool {
	if  c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" && 
	    c.Phone == "" && c.Document == "" {
		return true
	}
	return false
}
