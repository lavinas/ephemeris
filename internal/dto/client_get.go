package dto

import (
	"errors"
	"fmt"

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
func (c *ClientGetIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a string representation of the client
func (c *ClientGetIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewClient(c.ID, c.Date, c.Name, c.Email, c.Phone, c.Document, c.Contact),
	}
}

// GetOut is a method that returns the output dto
func (c *ClientGetIn) GetOut() port.DTOOut {
	return &ClientGetOut{}
}

// GetDTO is a method that returns the dto
func (c *ClientGetOut) GetDTO(domainIn interface{}) interface{} {
	ret := []ClientGetOut{}
	slices := domainIn.([]interface{})
	clients := slices[0].(*[]domain.Client)
	for _, client := range *clients {
		doc := ""
		if client.Document != nil {
			doc = *client.Document
		}
		dto := ClientGetOut{
			ID:       client.ID,
			Date:     client.Date.Format(pkg.DateFormat),
			Name:     client.Name,
			Email:    client.Email,
			Phone:    client.Phone,
			Document: doc,
			Contact:  client.Contact,
		}
		ret = append(ret, dto)
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

// IsEmpty is a method that returns true if the dto is empty
func (c *ClientGetIn) isEmpty() bool {
	fmt.Println(1, c)
	if c.ID == "" && c.Date == "" && c.Name == "" && c.Email == "" &&
		c.Phone == "" && c.Document == "" && c.Contact == "" {
		fmt.Println(2, c)
		return true
	}
	fmt.Println(3, c)
	return false
}
