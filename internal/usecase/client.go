package usecase

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
)

// Add is a method that add a client to the repository
func (c *Usecase) AddClient(id string, name string, responsible string, email string, 
	                        phone string, contactWay string, document string) error {
	c.Log.Println("Registering client")
	addMap := map[string]func(*domain.Client) error{
		"validate":       c.validateClient,
		"format":         c.formatClient,
		"checkExistence": c.checkExistsClient,
		"add":            c.addClient,
	}
	client := domain.NewClient(id, name, responsible, email, phone, contactWay, document)
	for _, f := range addMap {
		if err := f(client); err != nil {
			return err
		}
	}
	return nil
}

// Get is a method that gets a client from the repository
func (c *Usecase) GetClient(id string) (*domain.Client, error) {
	c.Log.Println("Getting client")
	client := &domain.Client{}
	if f, err := c.Repo.Get(client, id); err != nil {
		c.Log.Println(err.Error())
		return nil, errors.New("internal error: " + err.Error())
	} else if !f {
		err := errors.New("client not found")
		c.Log.Println(err.Error())
		return nil, errors.New("not found: " + err.Error())
	}
	return client, nil
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
