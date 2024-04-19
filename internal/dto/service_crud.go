package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ServiceCrud represents the dto for getting a service
type ServiceCrud struct {
	Object  string `json:"-" command:"name:service;key;pos:2-"`
	Action  string `json:"-" command:"name:get;key;pos:2-"`
	ID      string `json:"id" command:"name:id;pos:3+"`
	Date    string `json:"date" command:"name:date;pos:3+"`
	Name    string `json:"name" command:"name:name;pos:3+"`
	Minutes string `json:"minutes" command:"name:minutes;pos:3+"`
}

// Validate is a method that validates the dto
func (c *ServiceCrud) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *ServiceCrud) GetCommand() string {
	return p.Action
}


// GetDomain is a method that returns a string representation of the service
func (c *ServiceCrud) GetDomain() []port.Domain {
	if c.Action == "add" && c.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		c.Date = time.Now().Format(pkg.DateFormat)
	}
	if c.Action == "add" && c.Minutes == "" {
		c.Minutes = "0"
	}

	return []port.Domain{
		domain.NewService(c.ID, c.Date, c.Name, c.Minutes),
	}
}

// GetOut is a method that returns the output dto
func (c *ServiceCrud) GetOut() port.DTOOut {
	return &ServiceCrud{}
}

// GetDTO is a method that returns the dto
func (c *ServiceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	services := slices[0].(*[]domain.Service)
	for _, service := range *services {
		min := ""
		if service.Minutes != nil {
			min = strconv.FormatInt(*service.Minutes, 10)
		}
		dto := ServiceCrud{
			ID:      service.ID,
			Date:    service.Date.Format(pkg.DateFormat),
			Name:    service.Name,
			Minutes: min,
		}
		ret = append(ret, &dto)
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (c *ServiceCrud) isEmpty() bool {
	return c.ID == "" && c.Date == "" && c.Name == ""
}
