package dto

import (
	"github.com/lavinas/ephemeris/internal/domain"
)

// ClientAdd represents the dto for adding a client
type ClientAdd struct {
	Object      string `json:"-" command:"name:client;key"`
	Action      string `json:"-" command:"name:add;key"`
	ID          string `json:"id" command:"name:id"`
	Name        string `json:"name" command:"name:name"`
	Responsible string `json:"responsible" command:"name:responsible"`
	Email       string `json:"email" command:"name:email"`
	Phone       string `json:"phone" command:"name:phone"`
	Contact     string `json:"contact" command:"name:contact"`
	Document    string `json:"document" command:"name:document"`
}

// GetDomain is a method that returns a domain representation of the client dto
func (c *ClientAdd) GetDomain() *domain.Client {
	return domain.NewClient(c.ID, c.Name, c.Responsible, c.Email, c.Phone, c.Contact, c.Document)
}

// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientAdd) GetDto(domain *domain.Client) *ClientAdd {
	return &ClientAdd{
		ID:          domain.ID,
		Name:        domain.Name,
		Responsible: domain.Responsible,
		Email:       domain.Email,
		Phone:       domain.Phone,
		Contact:     domain.Contact,
		Document:    domain.Document,
	}
}

// ClientGet represents the dto for getting a client
type ClientGet struct {
	Object      string `json:"-" command:"name:client;key"`
	Action      string `json:"-" command:"name:get;key"`
	ID          string `json:"id" command:"name:id"`
	Name        string `json:"name" command:"name:name"`
	Responsible string `json:"responsible" command:"name:responsible"`
	Email       string `json:"email" command:"name:email"`
	Phone       string `json:"phone" command:"name:phone"`
	Contact     string `json:"contact" command:"name:contact"`
	Document    string `json:"document" command:"name:document"`
}

// GetDomain is a method that returns a string representation of the client
func (c *ClientGet) GetDomain() *domain.Client {
	return domain.NewClient(c.ID, c.Name, c.Responsible, c.Email, c.Phone, c.Contact, c.Document)
}

// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientGet) GetDto(domain *domain.Client) *ClientGet {
	return &ClientGet{
		ID:          domain.ID,
		Name:        domain.Name,
		Responsible: domain.Responsible,
		Email:       domain.Email,
		Phone:       domain.Phone,
		Contact:     domain.Contact,
		Document:    domain.Document,
	}
}
