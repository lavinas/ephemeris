package dto

import (
	// "errors"
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

// ClientAdd represents the input dto for adding a client usecase
type ClientAddIn struct {
	Object    string `json:"-" command:"name:client;key"`
	Action    string `json:"-" command:"name:add;key"`
	ID        string `json:"id" command:"name:id"`
	Date      string `json:"date" command:"name:date"`
	Name      string `json:"name" command:"name:name"`
	Email     string `json:"email" command:"name:email"`
	Phone     string `json:"phone" command:"name:phone"`
	Document  string `json:"document" command:"name:document"`
	Type      string `json:"type" command:"name:type"`
	Reference string `json:"reference" command:"name:reference"`
}

// ClientAddOut represents the output dto for adding a client usecase
type ClientAddOut struct {
	ID       string `json:"id" command:"name:id"`
	Date     string `json:"date" command:"name:date"`
	Name     string `json:"name" command:"name:name"`
	Email    string `json:"email" command:"name:email"`
	Phone    string `json:"phone" command:"name:phone"`
	Document string `json:"document" command:"name:document"`
}

// Validate is a method that validates the dto
func (c *ClientAddIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a domain representation of the client dto
func (c *ClientAddIn) GetDomain() []port.Domain {
	if c.Date == "" {
		time.Local, _ = time.LoadLocation(port.Location)
		c.Date = time.Now().Format(port.DateFormat)
	}
	roleId := fmt.Sprintf("%s_%s_%s", c.ID, port.RoleClient, c.ID)
	return []port.Domain{
		domain.NewClient(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, port.DefaultContact),
		domain.NewClientRole(roleId, c.Date, c.ID, port.RoleClient, c.ID),
	}
}

// SetDomain is a method that sets the dto with the domain
func (c *ClientAddOut) GetDTO(domainIn interface{}) interface{} {
	slices := domainIn.([]interface{})
	client := slices[0].(*domain.Client)
	dto := &ClientAddOut{}
	dto.ID = client.ID
	dto.Date = client.Date.Format(port.DateFormat)
	dto.Name = client.Name
	dto.Email = client.Email
	dto.Phone = client.Phone
	dto.Document = client.Document
	return dto
}

// isEmpty is a method that returns true if the dto is empty
func (c *ClientAddIn) isEmpty() bool {
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" {
		return true
	}
	return false
}
