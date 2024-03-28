package usecase

import (
	"errors"
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
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
func (c *Usecase) ClientAdd(in port.DTO) (interface{}, string, error) {
	dto, ok := in.(*dto.ClientAdd)
	if !ok {
		return nil, ErrWrongAddClientDTO, errors.New(ErrWrongAddClientDTO)
	}
	loop := []func(*domain.Client) error{
		c.validateClient,
		c.formatClient,
		c.checkExistsClient,
		c.addClient,
	}
	client := domain.NewClient(dto.ID, dto.Name, dto.Responsible, dto.Email,
		dto.Phone, dto.Contact, dto.Document)
	for _, f := range loop {
		if err := f(client); err != nil {
			return nil, err.Error(), err
		}
	}
	return nil, "ok: client added", nil
}

// Get is a method that gets a client from the repository
func (c *Usecase) ClientGet(dtoIn port.DTO) (interface{}, string, error) {
	din, ok := dtoIn.(*dto.ClientGet)
	if !ok {
		c.Log.Println(ErrWrongGetClientDTO)
		return nil, ErrWrongGetClientDTO, errors.New(ErrWrongGetClientDTO)
	}
	client := domain.NewClient(din.ID, din.Name, din.Responsible, din.Email,
		din.Phone, din.Contact, din.Document)
	client.Format()
	clients := []domain.Client{}
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
	dout := []*dto.ClientGet{}
	for _, client := range clients {
		d := &dto.ClientGet{}
		d.ID = client.ID
		d.Name = client.Name
		d.Responsible = client.Responsible
		d.Email = client.Email
		d.Phone = client.Phone
		d.Contact = client.Contact
		d.Document = client.Document
		dout = append(dout, d)
		strout += comm.MarshallNoKeys(d) + "\n"
	}
	return dout, strout, nil
}

// validate is a method that validates the client
func (c *Usecase) validateClient(client *domain.Client) error {
	if err := client.Validate(); err != nil {
		c.Log.Println(err.Error())
		return errors.New("bad request: " + err.Error())
	}
	return nil
}

// format is a method that formats the client
func (c *Usecase) formatClient(client *domain.Client) error {
	client.Format()
	return nil
}

// checkExistence is a method that checks if the client exists
func (c *Usecase) checkExistsClient(client *domain.Client) error {
	if f, err := c.Repo.Get(&domain.Client{}, client.ID); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	} else if f {
		err := fmt.Sprintf(ErrClientAlreadyExists, client.ID)
		c.Log.Println(err)
		return errors.New(err)
	}
	return nil
}

// add is a method that adds a client to the repository
func (c *Usecase) addClient(client *domain.Client) error {
	if err := c.Repo.Add(client); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	}
	return nil
}
