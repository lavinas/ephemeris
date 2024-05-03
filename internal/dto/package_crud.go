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
type PackageCrud struct {
	Object       string `json:"-" command:"name:package;key;pos:2-"`
	Action       string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID           string `json:"id" command:"name:id;pos:3+"`
	Date         string `json:"date" command:"name:date;pos:3+"`
	ServiceID    string `json:"service" command:"name:service;pos:3+"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence;pos:3+"`
	UnitValue    string `json:"unit" command:"name:unit;pos:3+"`
	PackValue    string `json:"pack" command:"name:pack;pos:3+"`
}

// Validate is a method that validates the dto
func (p *PackageCrud) Validate(repo port.Repository) error {
	if p.Action != "get" && p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *PackageCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageCrud) GetDomain() []port.Domain {
	if p.Action == "add" && p.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		p.Date = time.Now().Format(pkg.DateFormat)
	}
	return []port.Domain{
		domain.NewPackage(p.ID, p.Date, p.ServiceID, p.RecurrenceID, p.UnitValue, p.PackValue),
	}
}

// GetOut is a method that returns the output dto
func (p *PackageCrud) GetOut() port.DTOOut {
	return &PackageCrud{}
}

// GetDTO is a method that returns the dto
func (p *PackageCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	packages := slices[0].(*[]domain.Package)
	for _, p := range *packages {
		pack := ""
		if p.Price != nil {
			pack = fmt.Sprintf("%.2f", *p.Price)
		}
		for _, i := range p.Items {
			unit := ""
			if i.Price != nil {
				unit = fmt.Sprintf("%.2f", *i.Price)
			}
			ret = append(ret, &PackageCrud{
				ID:           p.ID,
				Date:         p.Date.Format(pkg.DateFormat),
				ServiceID:    i.ServiceID,
				RecurrenceID: p.RecurrenceID,
				UnitValue:    pack,
				PackValue:    unit,
			})
		}
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (p *PackageCrud) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.ServiceID == "" &&
		p.RecurrenceID == "" && p.UnitValue == "" && p.PackValue == ""
}
