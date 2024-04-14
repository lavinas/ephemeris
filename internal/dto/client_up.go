package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type ClientUpIn struct {
	Object   string `json:"-" command:"name:client;key;pos:2-"`
	Action   string `json:"-" command:"name:up;key;pos:2-"`
	ID       string `json:"id" command:"name:id;pos:3+"`
	Date     string `json:"date" command:"name:date;pos:3+"`
	Name     string `json:"name" command:"name:name;pos:3+"`
	Email    string `json:"email" command:"name:email;pos:3+"`
	Phone    string `json:"phone" command:"name:phone;pos:3+"`
	Document string `json:"document" command:"name:document;pos:3+"`
	Contact  string `json:"contact" command:"name:contact;pos:3+"`
}

// ClientUpOut represents the output dto for updating a client usecase
type ClientUpOut struct {
	ID       string `json:"id" command:"name:id"`
	Date     string `json:"date" command:"name:date"`
	Name     string `json:"name" command:"name:name"`
	Email    string `json:"email" command:"name:email"`
	Phone    string `json:"phone" command:"name:phone"`
	Document string `json:"document" command:"name:document"`
	Contact  string `json:"contact" command:"name:contact"`
}

// Validate is a method that validates the dto
func (c *ClientUpIn) Validate(repo port.Repository) error {
	if c.IsEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	if c.ID == "" {
		return errors.New(pkg.ErrIdUninformed)
	}
	id := c.ID
	c.ID = ""
	if c.IsEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	c.ID = id
	return nil
}

// GetDomain is a method that returns a domain representation of the client dto
func (c *ClientUpIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewClient(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, c.Contact),
	}
}

// GetOut is a method that returns the output dto
func (c *ClientUpIn) GetOut() port.DTOOut {
	return &ClientUpOut{}
}

// SetDomain is a method that sets the dto with the domain
func (c *ClientUpOut) GetDTO(domainIn interface{}) []port.DTOOut {
	slices := domainIn.([]interface{})
	client := slices[0].(*domain.Client)
	dto := &ClientUpOut{}
	dto.ID = client.ID
	dto.Date = client.Date.Format(pkg.DateFormat)
	dto.Name = client.Name
	dto.Email = client.Email
	dto.Phone = client.Phone
	dto.Document = ""
	if client.Document != nil {
		dto.Document = *client.Document
	}
	dto.Contact = client.Contact
	return []port.DTOOut{dto}
}

// IsEmpty is a method that returns true if the dto is empty
func (c *ClientUpIn) IsEmpty() bool {
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" && c.Contact == "" {
		return true
	}
	return false
}
