package usecase

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

// ClientUsecase is a struct that defines the usecase of the client
type ClientUsecase struct {
	Repo port.Repository
	Log  port.Logger
	Cfg  port.Config
}

// Insert is a method that inserts a client
func (c *ClientUsecase) Register(id string, name string, responsible string, email string, phone string, contactWay string, document string) error {
	c.Log.Info("Registering client")
	client := domain.NewClient(id, name, responsible, email, phone, contactWay, document)
	if err := client.Validate(); err != nil {
		c.Log.Info(err.Error())
		return errors.New("bad request: " + err.Error())
	}
	if err := client.Format(); err != nil {
		c.Log.Info(err.Error())
		return errors.New("internal error: " + err.Error())
	}
	if f, err := c.Repo.Get(&domain.Client{}, client.GetID()	); err != nil {
		c.Log.Info(err.Error())
		return errors.New("internal error: " + err.Error())
	} else if f {
		err := errors.New("client already exists")
		c.Log.Info(err.Error())
		return errors.New("conflict: " + err.Error())
	}
	if err := c.Repo.Add(client); err != nil {
		c.Log.Info(err.Error())
		return errors.New("internal error: " + err.Error())
	}
	return nil
}

// Get is a method that gets a client
func (c *ClientUsecase) Get(id string) (*domain.Client, error) {
	c.Log.Info("Getting client")
	client := &domain.Client{}
	if f, err := c.Repo.Get(client, id); err != nil {
		c.Log.Info(err.Error())
		return nil, errors.New("internal error: " + err.Error())
	} else if !f {
		err := errors.New("client not found")
		c.Log.Info(err.Error())
		return nil, errors.New("not found: " + err.Error())
	}
	return client, nil
}