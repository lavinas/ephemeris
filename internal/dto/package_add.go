package dto

import (
	"errors"
	"time"
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PackageAddIn represents the input dto for adding a package usecase
type PackageAddIn struct {
	Object       string `json:"-" command:"name:package;key;pos:1,2"`
	Action       string `json:"-" command:"name:add;key:pos:1,2"`
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	Name         string `json:"name" command:"name:name"`
	ServiceID    string `json:"service" command:"name:service"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence"`
	PriceID      string `json:"price" command:"name:price"`
}

// PackageAddOut represents the output dto for adding a package usecase
type PackageAddOut struct {
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	Name         string `json:"name" command:"name:name"`
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

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageAddIn) GetDomain() []port.Domain {
	if p.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		p.Date = time.Now().Format(pkg.DateFormat)
	}
	fmt.Println(1, p.ID, p.Date, p.Name, p.ServiceID, p.RecurrenceID, p.PriceID)
	return []port.Domain{
		domain.NewPackage(p.ID, p.Date, p.Name, p.ServiceID, p.RecurrenceID, p.PriceID),
	}
}

// GetOut is a method that returns the output dto
func (p *PackageAddIn) GetOut() port.DTOOut {
	return &PackageAddOut{}
}

// GetDTO is a method that returns the dto
func (p *PackageAddOut) GetDTO(domainIn interface{}) interface{} {
	{
		slices := domainIn.([]interface{})
		pg := slices[0].(*domain.Package)
		return &PackageAddOut{
			ID:           pg.ID,
			Date:         pg.Date.Format(pkg.DateFormat),
			Name:         pg.Name,
			ServiceID:    pg.ServiceID,
			RecurrenceID: pg.RecurrenceID,
			PriceID:      pg.PriceID,
		}
	}
}

// isEmpty is a method that checks if the dto is empty
func (p *PackageAddIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.Name == "" && p.ServiceID == "" &&
		p.RecurrenceID == "" && p.PriceID == ""
}
