package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PackageUpIn represents the input dto for updating a package usecase
type PackageUpIn struct {
	Object       string `json:"-" command:"name:package;key;pos:2-"`
	Action       string `json:"-" command:"name:up;key;pos:2-"`
	ID           string `json:"id" command:"name:id;pos:3+"`
	Date         string `json:"date" command:"name:date;pos:3+"`
	ServiceID    string `json:"service" command:"name:service;pos:3+"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence;pos:3+"`
	PriceID      string `json:"price" command:"name:price;pos:3+"`
}

// PackageUpOut represents the output dto for updating a package usecase
type PackageUpOut struct {
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	ServiceID    string `json:"service" command:"name:service"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence"`
	PriceID      string `json:"price" command:"name:price"`
}

// Validate is a method that validates the dto
func (p *PackageUpIn) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageUpIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewPackage(p.ID, p.Date, p.ServiceID, p.RecurrenceID, p.PriceID),
	}
}

// GetOut is a method that returns the output dto
func (p *PackageUpIn) GetOut() port.DTOOut {
	return &PackageUpOut{}
}

// GetDTO is a method that returns the dto
func (p *PackageUpOut) GetDTO(domainIn interface{}) []port.DTOOut {
	slices := domainIn.([]interface{})
	pg := slices[0].(*domain.Package)
	return []port.DTOOut{&PackageUpOut{
		ID:           pg.ID,
		Date:         pg.Date.Format(pkg.DateFormat),
		ServiceID:    pg.ServiceID,
		RecurrenceID: pg.RecurrenceID,
		PriceID:      pg.PriceID,
	}}
}

// isEmpty is a method that checks if the dto is empty
func (p *PackageUpIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.ServiceID == "" && p.RecurrenceID == "" && p.PriceID == ""
}

