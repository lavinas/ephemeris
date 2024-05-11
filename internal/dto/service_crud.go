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
	Action  string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort    string `json:"sort" command:"name:sort;pos:3+"`
	ID      string `json:"id" command:"name:id;pos:3+;trans:id,string"`
	Date    string `json:"date" command:"name:date;pos:3+;trans:date,time"`
	Name    string `json:"name" command:"name:name;pos:3+;trans:name,string"`
	Minutes string `json:"minutes" command:"name:minutes;pos:3+;trans:minutes,numeric"`
}

// Validate is a method that validates the dto
func (s *ServiceCrud) Validate(repo port.Repository) error {
	if s.Action != "get" && s.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *ServiceCrud) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the service
func (s *ServiceCrud) GetDomain() []port.Domain {
	if s.Action == "add" && s.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		s.Date = time.Now().Format(pkg.DateFormat)
	}
	if s.Action == "add" && s.Minutes == "" {
		s.Minutes = "0"
	}

	return []port.Domain{
		domain.NewService(s.ID, s.Date, s.Name, s.Minutes),
	}
}

// GetOut is a method that returns the output dto
func (s *ServiceCrud) GetOut() port.DTOOut {
	return s
}

// GetDTO is a method that returns the dto
func (s *ServiceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
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
	pkg.NewCommands().Sort(ret, s.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (s *ServiceCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	cmd, err := pkg.NewCommands().Transpose(s)
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) > 0 {
		domain := s.GetDomain()[0]
		return domain, cmd, nil
	}
	return domain, cmd, nil
}


// isEmpty is a method that checks if the dto is empty
func (s *ServiceCrud) isEmpty() bool {
	return s.ID == "" && s.Date == "" && s.Name == "" && s.Minutes == ""
}
