package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ClientGet represents the dto for getting a client
type ClientCrud struct {
	Object   string `json:"-" command:"name:client;key;pos:2-"`
	Action   string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID       string `json:"id" command:"name:id;;pos:3+"`
	Date     string `json:"date" command:"name:date;pos:3+"`
	Name     string `json:"name" command:"name:name;pos:3+"`
	Email    string `json:"email" command:"name:email;pos:3+"`
	Phone    string `json:"phone" command:"name:phone;pos:3+"`
	Document string `json:"document" command:"name:document;pos:3+"`
	Contact  string `json:"contact" command:"name:contact;pos:3+"`
}

// Validate is a method that validates the dto
func (c *ClientCrud) Validate(repo port.Repository) error {
	if c.Action != "get" && c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *ClientCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns a string representation of the client
func (c *ClientCrud) GetDomain() []port.Domain {
	if c.Action == "add" && c.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		c.Date = time.Now().Format(pkg.DateFormat)
	}
	if c.Action == "add" && c.Contact == "" {
		c.Contact = pkg.DefaultContact
	}
	return []port.Domain{
		domain.NewClient(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, c.Contact),
	}
}

// GetOut is a method that returns the output dto
func (c *ClientCrud) GetOut() port.DTOOut {
	return &ClientCrud{}
}

// GetDTO is a method that returns the dto
func (c *ClientCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	clients := slices[0].(*[]domain.Client)
	for _, client := range *clients {
		doc := ""
		if client.Document != nil {
			doc = *client.Document
		}
		dto := ClientCrud{
			ID:       client.ID,
			Date:     client.Date.Format(pkg.DateFormat),
			Name:     client.Name,
			Email:    client.Email,
			Phone:    client.Phone,
			Document: doc,
			Contact:  client.Contact,
		}
		ret = append(ret, &dto)
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

// IsEmpty is a method that returns true if the dto is empty
func (c *ClientCrud) isEmpty() bool {
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" && c.Contact == "" {
		return true
	}
	return false
}
