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
func (c *Usecase) ClientAdd(dtoIn port.DTO) (string, error) {
	dto, ok := dtoIn.(*dto.ClientAdd)
	if !ok {
		return ErrWrongAddClientDTO, errors.New(ErrWrongAddClientDTO)
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
			return err.Error(), err
		}
	}
	return "ok: client added", nil
}

// Get is a method that gets a client from the repository
func (c *Usecase) ClientGet(dtoIn port.DTO) (string, error) {
	dto, ok := dtoIn.(*dto.ClientGet)
	if !ok {
		c.Log.Println(ErrWrongGetClientDTO)
		return ErrWrongGetClientDTO, errors.New(ErrWrongGetClientDTO)
	}
	client := domain.NewClient(dto.ID, dto.Name, dto.Responsible, dto.Email,
		dto.Phone, dto.Contact, dto.Document)
	client.Format()
	fmt.Println(1, client)
	if f, err := c.Repo.Find(client); err != nil {
		errMsg := "internal error: " + err.Error()
		c.Log.Println(errMsg)
		return errMsg, errors.New(errMsg)
	} else if !f {
		errMsg := "not found: client not found"
		c.Log.Println(errMsg)
		return errMsg, errors.New(errMsg)
	}

	/*
		client := &domain.Client{}
		if f, err := c.Repo.Get(client, dto.ID); err != nil {
			errMsg := "internal error: " + err.Error()
			c.Log.Println(errMsg)
			return errMsg, errors.New(errMsg)
		} else if !f {
			errMsg := "not found: client not found"
			c.Log.Println(errMsg)
			return errMsg, errors.New(errMsg)
		}
	*/
	dto.ID = client.ID
	dto.Name = client.Name
	dto.Responsible = client.Responsible
	dto.Email = client.Email
	dto.Phone = client.Phone
	dto.Contact = client.Contact
	dto.Document = client.Document
	comm := &pkg.Commands{}
	return comm.MarshallNoKeys(dto), nil
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
		err := "conflict: " + fmt.Sprintf(ErrClientAlreadyExists, client.ID)
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
