package dto

import (
	"errors"
	"strconv"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

// ServiceGetIn represents the dto for getting a service
type ServiceGetIn struct {
	Object  string `json:"-" command:"name:service;key"`
	Action  string `json:"-" command:"name:get;key"`
	ID      string `json:"id" command:"name:id"`
	Date    string `json:"date" command:"name:date"`
	Name    string `json:"name" command:"name:name"`
	Minutes string `json:"minutes" command:"name:minutes"`
}

// ServiceGetOut represents the output dto for getting a service
type ServiceGetOut struct {
	ID      string `json:"id" command:"name:id"`
	Date    string `json:"date" command:"name:date"`
	Name    string `json:"name" command:"name:name"`
	Minutes string `json:"minutes" command:"name:minutes"`
}

// Validate is a method that validates the dto
func (c *ServiceGetIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a string representation of the service
func (c *ServiceGetIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewService(c.ID, c.Date, c.Name, c.Minutes),
	}
}

// GetOut is a method that returns the output dto
func (c *ServiceGetIn) GetOut() port.DTOOut {
	return &ServiceGetOut{}
}

// GetDTO is a method that returns the dto
func (c *ServiceGetOut) GetDTO(domainIn interface{}) interface{} {
	ret := []ServiceGetOut{}
	slices := domainIn.([]interface{})
	services := slices[0].(*[]domain.Service)
	for _, service := range *services {
		dto := ServiceGetOut{
			ID:      service.ID,
			Date:    service.Date.Format(port.DateFormat),
			Name:    service.Name,
			Minutes: strconv.FormatInt(service.Minutes, 10),
		}
		ret = append(ret, dto)
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (c *ServiceGetIn) isEmpty() bool {
	return c.ID == "" && c.Date == "" && c.Name == ""
}
