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
	Base
	Object   string `json:"-" command:"name:client;key;pos:2-"`
	Action   string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort     string `json:"sort" command:"name:sort;pos:3+"`
	Csv      string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID       string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date     string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	Name     string `json:"name" command:"name:name;pos:3+;trans:name,string" csv:"name"`
	Email    string `json:"email" command:"name:email;pos:3+;trans:email,string" csv:"email"`
	Phone    string `json:"phone" command:"name:phone;pos:3+;trans:phone,string" csv:"phone"`
	Document string `json:"document" command:"name:document;pos:3+;trans:document,string" csv:"document"`
	Contact  string `json:"contact" command:"name:contact;pos:3+;trans:contact,string" csv:"contact"`
}

// Validate is a method that validates the dto
func (c *ClientCrud) Validate(repo port.Repository) error {
	if c.Csv != "" && (c.ID != "" || c.Date != "" || c.Name != "" || c.Email != "" || c.Phone != "" || c.Document != "" || c.Contact != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *ClientCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns a string representation of the client
func (c *ClientCrud) GetDomain() []port.Domain {
	if c.Csv != "" {
		domains := []port.Domain{}
		clients := []*ClientCrud{}
		c.ReadCSV(&clients, c.Csv)
		for _, client := range clients {
			client.Action = c.Action
			client.Object = c.Object
			domains = append(domains, c.getDomain(client))
		}
		return domains
	}
	return []port.Domain{c.getDomain(c)}
}

// getDomain is a method that returns a string representation of the agenda
func (c *ClientCrud) getDomain(one *ClientCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Contact == "" {
		one.Contact = pkg.DefaultContact
	}
	return domain.NewClient(one.ID, one.Date, one.Name, one.Email, one.Phone, one.Document, one.Contact)
}


// GetOut is a method that returns the output dto
func (c *ClientCrud) GetOut() port.DTOOut {
	return c
}

// GetDTO is a method that returns the dto
func (c *ClientCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		clients := slice.(*[]domain.Client)
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
	}
	pkg.NewCommands().Sort(ret, c.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (c *ClientCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return c.getInstructions(c, domain)
}

