package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PackageGetIn represents the input dto for getting a package usecase
type PackageGetIn struct {
	Object string `json:"-" command:"name:package;key;pos:2-"`
	Action string `json:"-" command:"name:get;key;pos:2-"`
	ID     string `json:"id" command:"name:id;pos:3+"`
	Date         string `json:"date" command:"name:date;pos:3+"`
	ServiceID    string `json:"service" command:"name:service;pos:3+"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence;pos:3+"`
	PriceID      string `json:"price" command:"name:price;pos:3+"`
}

// PackageGetOut represents the output dto for getting a package usecase
type PackageGetOut struct {
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	ServiceID    string `json:"service" command:"name:service"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence"`
	PriceID      string `json:"price" command:"name:price"`
}

// Validate is a method that validates the dto
func (p *PackageGetIn) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageGetIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewPackage(p.ID, p.Date, p.ServiceID, p.RecurrenceID, p.PriceID),
	}
}

// GetOut is a method that returns the output dto
func (p *PackageGetIn) GetOut() port.DTOOut {
	return &PackageGetOut{}
}

// GetDTO is a method that returns the dto
func (p *PackageGetOut) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	packages := slices[0].(*[]domain.Package)
	for _, p := range *packages {
		ret = append(ret, &PackageGetOut{
			ID:           p.ID,
			Date:         p.Date.Format(pkg.DateFormat),
			ServiceID:    p.ServiceID,
			RecurrenceID: p.RecurrenceID,
			PriceID:      p.PriceID,
		})
	}
	return ret
}

func (p *PackageGetIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.ServiceID == "" && p.RecurrenceID == "" && p.PriceID == ""
}