package usecase

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
)

const (
	ErrWrongAddClientDTO = "internal error: wrong AddClient dto"
	ErrWrongGetClientDTO = "internal error: wrong GetClient dto"
)

// Add is a method that add a client to the repository
func (c *Usecase) AddClient(dtoIn port.DTO) error {
	dto, ok := dtoIn.(*dto.ClientAdd)
	if !ok {
		return errors.New(ErrWrongAddClientDTO)
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
			return err
		}
	}
	return nil
}

// Get is a method that gets a client from the repository
func (c *Usecase) GetClient(dtoIn port.DTO) error {
	dto, ok := dtoIn.(*dto.ClientGet)
	if !ok {
		return errors.New(ErrWrongGetClientDTO)
	}
	client := &domain.Client{}
	if f, err := c.Repo.Get(client, dto.ID); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	} else if !f {
		err := errors.New("client not found")
		c.Log.Println(err.Error())
		return errors.New("not found: " + err.Error())
	}
	dto.ID = client.ID
	dto.Name = client.Name
	dto.Responsible = client.Responsible
	dto.Email = client.Email
	dto.Phone = client.Phone
	dto.Contact = client.Contact
	dto.Document = client.Document
	return nil
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
	if err := client.Format(); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	}
	return nil
}

// checkExistence is a method that checks if the client exists
func (c *Usecase) checkExistsClient(client *domain.Client) error {
	if f, err := c.Repo.Get(&domain.Client{}, client.ID); err != nil {
		c.Log.Println(err.Error())
		return errors.New("internal error: " + err.Error())
	} else if f {
		err := errors.New("client already exists")
		c.Log.Println(err.Error())
		return errors.New("conflict: " + err.Error())
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
