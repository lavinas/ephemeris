package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PackageAddIn represents the input dto for adding a package usecase
type PackageAddIn struct {
	Object       string `json:"-" command:"name:package;key;pos:2-"`
	Action       string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID           string `json:"id" command:"name:id;pos:3+"`
	Date         string `json:"date" command:"name:date;pos:3+"`
	ServiceID    string `json:"service" command:"name:service;pos:3+"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence;pos:3+"`
	PriceID      string `json:"price" command:"name:price;pos:3+"`
}

// PackageAddOut represents the output dto for adding a package usecase
type PackageAddOut struct {
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	ServiceID    string `json:"service" command:"name:service"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence"`
	PriceID      string `json:"price" command:"name:price"`
}

// Validate is a method that validates the dto
func (p *PackageAddIn) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *PackageAddIn) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageAddIn) GetDomain() []port.Domain {
	if p.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		p.Date = time.Now().Format(pkg.DateFormat)
	}
	return []port.Domain{
		domain.NewPackage(p.ID, p.Date, p.ServiceID, p.RecurrenceID, p.PriceID),
	}
}

// GetOut is a method that returns the output dto
func (p *PackageAddIn) GetOut() port.DTOOut {
	return &PackageAddOut{}
}

// GetDTO is a method that returns the dto
/*
func (p *PackageAddOut) GetDTO(domainIn interface{}) []port.DTOOut {
	{
		slices := domainIn.([]interface{})
		pg := slices[0].(*domain.Package)
		return []port.DTOOut{&PackageAddOut{
			ID:           pg.ID,
			Date:         pg.Date.Format(pkg.DateFormat),
			ServiceID:    pg.ServiceID,
			RecurrenceID: pg.RecurrenceID,
			PriceID:      pg.PriceID,
		}}
	}
}
*/
func (p *PackageAddOut) GetDTO(domainIn interface{}) []port.DTOOut {
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

// isEmpty is a method that checks if the dto is empty
func (p *PackageAddIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.ServiceID == "" &&
		p.RecurrenceID == "" && p.PriceID == ""
}
