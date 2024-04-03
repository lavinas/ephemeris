package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

// ClientGet represents the dto for getting a client
type ClientGetIn struct {
	Object   string `json:"-" command:"name:client;key"`
	Action   string `json:"-" command:"name:get;key"`
	ID       string `json:"id" command:"name:id"`
	Date     string `json:"date" command:"name:date"`
	Name     string `json:"name" command:"name:name"`
	Email    string `json:"email" command:"name:email"`
	Phone    string `json:"phone" command:"name:phone"`
	Document string `json:"document" command:"name:document"`
	Contact  string `json:"contact" command:"name:contact"`
}

// ClientGetOut represents the output dto for getting a client
type ClientGetOut struct {
	ID       string `json:"id" command:"name:id"`
	Date     string `json:"date" command:"name:date"`
	Name     string `json:"name" command:"name:name"`
	Email    string `json:"email" command:"name:email"`
	Phone    string `json:"phone" command:"name:phone"`
	Document string `json:"document" command:"name:document"`
	Contact  string `json:"contact" command:"name:contact"`
}

// Validate is a method that validates the dto
func (c *ClientGetIn) Validate() error {
	if c.isEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a string representation of the client
func (c *ClientGetIn) GetDomain() port.Domain {
	return domain.NewClient(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, c.Contact)
}

// GetDTO is a method that returns the dto
func (c *ClientGetOut) GetDTO(domainIn interface{}) interface{} {
	ret := []ClientGetOut{}
	din := domainIn.(*[]domain.Client)
	for _, d := range *din {
		dto := ClientGetOut{
			ID:       d.ID,
			Date:     d.Date.Format(port.DateFormat),
			Name:     d.Name,
			Email:    d.Email,
			Phone:    d.Phone,
			Document: d.Document,
			Contact:  d.Contact,
		}
		ret = append(ret, dto)
	}
	return ret
}

// IsEmpty is a method that returns true if the dto is empty
func (c *ClientGetIn) isEmpty() bool {
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" {
		return true
	}
	return false
}
