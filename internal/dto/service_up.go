package dto

import (
	"errors"
	"strconv"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ServiceUpIn is a struct that represents the service up data transfer object
type ServiceUpIn struct {
	Object  string `json:"-" command:"name:service;key"`
	Action  string `json:"-" command:"name:up;key"`
	ID      string `json:"id" command:"name:id"`
	Date    string `json:"date" command:"name:date"`
	Name    string `json:"name" command:"name:name"`
	Minutes string `json:"minutes" command:"name:minutes"`
}

// ServiceUpOut is a struct that represents the service up output data transfer object
type ServiceUpOut struct {
	ID      string `json:"id" command:"name:id"`
	Date    string `json:"date" command:"name:date"`
	Name    string `json:"name" command:"name:name"`
	Minutes string `json:"minutes" command:"name:minutes"`
}

// Validate is a method that validates the dto
func (c *ServiceUpIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	if c.ID == "" {
		return errors.New(pkg.ErrIdUninformed)
	}
	id := c.ID
	c.ID = ""
	if c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	c.ID = id
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (c *ServiceUpIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewService(c.ID, c.Date, c.Name, c.Minutes),
	}
}

// GetOut is a method that returns the dto out
func (c *ServiceUpIn) GetOut() port.DTOOut {
	return &ServiceUpOut{}
}

// GetDTO is a method that returns the dto
func (c *ServiceUpOut) GetDTO(domainIn interface{}) interface{} {
	slices := domainIn.([]interface{})
	service, ok := slices[0].(*domain.Service)
	if !ok {
		return nil
	}
	min := ""
	if service.Minutes != nil {
		min = strconv.FormatInt(*service.Minutes, 10)
	}
	return &ServiceUpOut{
		ID:      service.ID,
		Date:    service.Date.Format(pkg.DateFormat),
		Name:    service.Name,
		Minutes: min,
	}
}

// GetDomain is a method that returns the domain of the dto
func (c *ServiceUpIn) isEmpty() bool {
	return c.ID == "" && c.Date == "" && c.Name == "" && c.Minutes == ""
}
