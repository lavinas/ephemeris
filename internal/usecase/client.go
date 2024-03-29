package usecase

import (
	"errors"
	"fmt"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	ErrWrongAddClientDTO   = "internal error: wrong AddClient dto"
	ErrWrongGetClientDTO   = "internal error: wrong GetClient dto"
	ErrClientAlreadyExists = "conflict: client already exists with id %s"
)

// Add is a method that add a client to the repository
func (c *Usecase) Add(in interface{}) (interface{}, string, error) {
	dto, ok := in.(*dto.ClientAdd)
	if !ok {
		c.Log.Println(ErrWrongAddClientDTO)
		return nil, ErrWrongAddClientDTO, errors.New(ErrWrongAddClientDTO)
	}
	loop := []func(domain port.Domain) error{
		c.validate,
		c.format,
		c.checkExists,
		c.addClient,
	}
	client := dto.GetDomain()
	for _, f := range loop {
		if err := f(client); err != nil {
			return nil, err.Error(), err
		}
	}
	return nil, "ok: client added", nil
}

// Get is a method that gets a client from the repository
func (c *Usecase) ClientGet(in interface{}) (interface{}, string, error) {
	din, ok := in.(*dto.ClientGet)
	if !ok {
		c.Log.Println(ErrWrongGetClientDTO)
		return nil, ErrWrongGetClientDTO, errors.New(ErrWrongGetClientDTO)
	}
	client := din.GetDomain()
	client.Format()
	clients := []port.Domain{}
	if err := c.Repo.Find(client, &clients); err != nil {
		errMsg := "internal error: " + err.Error()
		c.Log.Println(errMsg)
		return nil, errMsg, errors.New(errMsg)
	}
	if len(clients) == 0 {
		errMsg := fmt.Sprintf("not found: client with id %s", din.ID)
		c.Log.Println(errMsg)
		return nil, errMsg, errors.New(errMsg)
	}
	strout := ""
	comm := &pkg.Commands{}
	out := []*dto.ClientGet{}
	for _, client := range clients {
		dout := din.GetDto(&client)
		out = append(out, dout) 
		strout += comm.MarshallNoKeys(dout) + "\n"
	}
	return out, strout, nil
}

// validate is a method that validates the client
func (c *Usecase) validate(domain port.Domain) error {
	if err := domain.Validate(); err != nil {
		c.Log.Println(err.Error())
		return errors.New("bad request: " + err.Error())
	}
	return nil
}

// format is a method that formats the client
func (c *Usecase) format(domain port.Domain) error {
	domain.Format()
	return nil
}

// checkExistence is a method that checks if the client exists
func (c *Usecase) checkExists(domain port.Domain) error {

	if f, err := c.Repo.Get(domain, domain.GetID()); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	} else if f {
		err := fmt.Sprintf(ErrClientAlreadyExists, domain.GetID())
		c.Log.Println(err)
		return errors.New(err)
	}
	return nil
}

// add is a method that adds a client to the repository
func (c *Usecase) addClient(domain port.Domain) error {
	if err := c.Repo.Add(domain); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	}
	return nil
}

