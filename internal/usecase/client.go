package usecase

import (
	"errors"
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
)

const (
	ErrorClientCommandShort = "client command should have at least 1 parameter"
)


var (
	/*
	cmdsClient = map[string]func(*Usecase, string) string{
		"add": (*Usecase).CommandAddClient,
		"get": (*Usecase).CommandClientGet,
	}
	*/
)

// CommandClient is a method that receives a command and execute it
func (c *Usecase) CommandClient(cmd string) string {
	cmd = strings.ToLower(cmd)
	cmdSlice := strings.Split(cmd, " ")
	if len(cmdSlice) == 0 {
		return ErrorClientCommandShort
	}
	return "ok"
}

// CommandAddClient is a method that receives a command and execute it
func (c *Usecase) CommandAddClient(cmd string) string {
	cmd = strings.ToLower(cmd)
	strings.Split(cmd, " ")
	return "ok"
}
		
// Add is a method that add a client to the repository
func (c *Usecase) AddClient(dto *dto.ClientAdd) error {
	addSlice := []func(*domain.Client) error{
		c.validateClient,
		c.formatClient,
		c.checkExistsClient,
		c.addClient,
	}
	client := domain.NewClient(dto.ID, dto.Name, dto.Responsible, dto.Email, 
		                       dto.Phone, dto.Contact, dto.Document)
	c.Log.Println("Doc1: " + client.Document)
	for _, f := range addSlice {
		if err := f(client); err != nil {
			return err
		}
	}
	return nil
}

// Get is a method that gets a client from the repository
func (c *Usecase) GetClient(id string) (*dto.ClientGet, error) {
	client := &domain.Client{}
	if f, err := c.Repo.Get(client, id); err != nil {
		c.Log.Println(err.Error())
		return nil, errors.New("internal error: " + err.Error())
	} else if !f {
		err := errors.New("client not found")
		c.Log.Println(err.Error())
		return nil, errors.New("not found: " + err.Error())
	}
	dto := &dto.ClientGet{ID: client.ID, Name: client.Name, Responsible: client.Responsible,
		Email: client.Email,  Phone: client.Phone, Contact: client.Contact, Document: client.Document,
	}
	return dto, nil
}

// validate is a method that validates the client
func (c *Usecase) validateClient(client *domain.Client) error {
	c.Log.Println("Doc: " + client.Document)
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
	if f, err := c.Repo.Get(&domain.Client{}, client.GetID()); err != nil {
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
