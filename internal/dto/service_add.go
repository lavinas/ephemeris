package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ServiceAdd is a struct that represents the service add data transfer object
type ServiceAddIn struct {
	Object  string `json:"-" command:"name:service;key;pos:2-"`
	Action  string `json:"-" command:"name:add;key;pos:2-"`
	ID      string `json:"id" command:"name:id;pos:3+"`
	Date    string `json:"date" command:"name:date;pos:3+"`
	Name    string `json:"name" command:"name:name;pos:3+"`
	Minutes string `json:"minutes" command:"name:minutes;pos:3+"`
}

// ServiceAddOut is a struct that represents the service add output data transfer object
type ServiceAddOut struct {
	ID      string `json:"id" command:"name:id"`
	Date    string `json:"date" command:"name:date"`
	Name    string `json:"name" command:"name:name"`
	Minutes string `json:"minutes" command:"name:minutes"`
}

// Validate is a method that validates the dto
func (c *ServiceAddIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (c *ServiceAddIn) GetDomain() []port.Domain {
	if c.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		c.Date = time.Now().Format(pkg.DateFormat)
	}
	if c.Minutes == "" {
		c.Minutes = "0"
	}
	return []port.Domain{
		domain.NewService(c.ID, c.Date, c.Name, c.Minutes),
	}
}

// GetOut is a method that returns the dto out
func (c *ServiceAddIn) GetOut() port.DTOOut {
	return &ServiceAddOut{}
}

// GetDTO is a method that returns the dto
func (c *ServiceAddOut) GetDTO(domainIn interface{}) []port.DTOOut {
	slices := domainIn.([]interface{})
	service, ok := slices[0].(*domain.Service)
	if !ok {
		return nil
	}
	min := ""
	if service.Minutes != nil {
		min = strconv.FormatInt(*service.Minutes, 10)
	}
	return []port.DTOOut{&ServiceAddOut{
		ID:      service.ID,
		Date:    service.Date.Format(pkg.DateFormat),
		Name:    service.Name,
		Minutes: min,
	}}
}

// GetDomain is a method that returns the domain of the dto
func (c *ServiceAddIn) isEmpty() bool {
	return c.ID == "" && c.Date == "" && c.Name == "" && c.Minutes == ""
}
