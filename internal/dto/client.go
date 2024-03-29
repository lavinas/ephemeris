package dto

import (
	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	
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
func (c *ClientAdd) GetDomain() port.Domain {
	return domain.NewClient(c.ID, c.Name, c.Responsible, c.Email, c.Phone, c.Contact, c.Document)
}

// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientAdd) GetDto(in interface{}) interface{} {
	d := in.(*domain.Client)
	return &ClientAdd{
		ID:          d.ID,
		Name:        d.Name,
		Responsible: d.Responsible,
		Email:       d.Email,
		Phone:       d.Phone,
		Contact:     d.Contact,
		Document:    d.Document,
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
func (c *ClientGet) GetDomain() port.Domain {
	return domain.NewClient(c.ID, c.Name, c.Responsible, c.Email, c.Phone, c.Contact, c.Document)
}


// GetDto is a method that returns a DTO representation of the client domain
func (c *ClientGet) GetDto(in interface{}) interface{} {
	ret := make([]ClientGet, 0)
	d := in.(*[]domain.Client)
	for _, v := range *d {
		ret = append(ret, ClientGet{
			ID:          v.ID,
			Name:        v.Name,
			Responsible: v.Responsible,
			Email:       v.Email,
			Phone:       v.Phone,
			Contact:     v.Contact,
			Document:    v.Document,
		})
	}
	return ret
}

/*
func (c *ClientGet) GetDto(in interface{}) interface{} {
	ret := ClientGet{}
	d := in.(*[]domain.Client)
	if len(*d) == 0 {
		return ret
	}
	ret = ClientGet{
		ID:          (*d)[0].ID,
		Name:        (*d)[0].Name,
		Responsible: (*d)[0].Responsible,
		Email:       (*d)[0].Email,
		Phone:       (*d)[0].Phone,
		Contact:     (*d)[0].Contact,
		Document:    (*d)[0].Document,
	}
	return &ret
}
*/