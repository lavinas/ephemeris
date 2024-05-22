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
	Base
	Object  string `json:"-" command:"name:service;key;pos:2-"`
	Action  string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort    string `json:"sort" command:"name:sort;pos:3+"`
	Csv     string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID      string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date    string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	Name    string `json:"name" command:"name:name;pos:3+;trans:name,string" csv:"name"`
	Minutes string `json:"minutes" command:"name:minutes;pos:3+;trans:minutes,numeric" csv:"minutes"`
}

// Validate is a method that validates the dto
func (s *ServiceCrud) Validate(repo port.Repository) error {
	if s.Csv != "" && (s.ID != "" || s.Date != "" || s.Name != "" || s.Minutes != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *ServiceCrud) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the service
func (s *ServiceCrud) GetDomain() []port.Domain {
	if s.Csv != "" {
		domains := []port.Domain{}
		services := []*ServiceCrud{}
		s.ReadCSV(&services, s.Csv)
		for _, service := range services {
			service.Action = s.Action
			service.Object = s.Object
			domains = append(domains, s.getDomain(service))
		}
		return domains
	}
	return []port.Domain{s.getDomain(s)}
}

// getDomain is a method that returns a string representation of the service
func (s *ServiceCrud) getDomain(one *ServiceCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Minutes == "" {
		one.Minutes = "0"
	}
	return domain.NewService(one.ID, one.Date, one.Name, one.Minutes)
}

// GetOut is a method that returns the output dto
func (s *ServiceCrud) GetOut() port.DTOOut {
	return s
}

// GetDTO is a method that returns the dto
func (s *ServiceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		services := slice.(*[]domain.Service)
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
	}
	pkg.NewCommands().Sort(ret, s.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (s *ServiceCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return s.getInstructions(s, domain)
}
