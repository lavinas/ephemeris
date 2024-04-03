package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
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

// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientGetIn) GetOut(in interface{}) ([]port.DTOOut, string) {
	d := in.(*[]domain.Client)
	time.Local, _ = time.LoadLocation("America/Sao_Paulo")
	ret := []port.DTOOut{}
	for _, v := range *d {
		ret = append(ret, ClientGetIn{
			ID:       v.ID,
			Name:     v.Name,
			Date:     v.Date.Format(port.DateFormat),
			Email:    v.Email,
			Phone:    v.Phone,
			Document: v.Document,
			Contact:  v.Contact,
		})
	}
	if len(ret) == 0 {
		return nil, ""
	}
	return ret, pkg.NewCommands().Marshal(ret, "nokeys")
}

// IsEmpty is a method that returns true if the dto is empty
func (c *ClientGetIn) isEmpty() bool {
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" {
		return true
	}
	return false
}
