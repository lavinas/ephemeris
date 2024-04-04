package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

type ClientUpIn struct {
	Object   string `json:"-" command:"name:client;key"`
	Action   string `json:"-" command:"name:up;key"`
	ID       string `json:"id" command:"name:id"`
	Date     string `json:"contact" command:"name:date"`
	Name     string `json:"name" command:"name:name"`
	Email    string `json:"email" command:"name:email"`
	Phone    string `json:"phone" command:"name:phone"`
	Document string `json:"document" command:"name:document"`
}

// ClientUpOut represents the output dto for updating a client usecase
type ClientUpOut struct {
	ID       string `json:"id" command:"name:id"`
	Date     string `json:"date" command:"name:date"`
	Name     string `json:"name" command:"name:name"`
	Email    string `json:"email" command:"name:email"`
	Phone    string `json:"phone" command:"name:phone"`
	Document string `json:"document" command:"name:document"`
}

// GetDomain is a method that returns a domain representation of the client dto
func (c *ClientUpIn) GetDomain() []port.Domain {
	return []port.Domain{domain.NewClient(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, "")}
}

// SetDomain is a method that sets the dto with the domain
func (c *ClientUpOut) GetDTO(domainIn interface{}) interface{} {
	dto := &ClientUpOut{}
	domain := domainIn.(*domain.Client)
	dto.ID = domain.ID
	dto.Date = domain.Date.Format(port.DateFormat)
	dto.Name = domain.Name
	dto.Email = domain.Email
	dto.Phone = domain.Phone
	dto.Document = domain.Document
	return dto
}

// Validate is a method that validates the dto
func (c *ClientUpIn) Validate() error {
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
func (c *ClientUpIn) IsEmpty() bool {
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" {
		return true
	}
	return false
}
