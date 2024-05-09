package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PriceCrud is a struct that represents the price get data transfer object
type PriceCrud struct {
	Object string `json:"-" command:"name:price;key;pos:2-"`
	Action string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID     string `json:"id" command:"name:id;pos:3+;trans:string"`
	Date   string `json:"date" command:"name:date;pos:3+;trans:time"`
	Name   string `json:"name" command:"name:name;pos:3+;trans:string"`
	Unit   string `json:"unit" command:"name:unit;pos:3+;trans:numeric"`
	Pack   string `json:"pack" command:"name:pack;pos:3+;trans:numeric"`
}

// Validate is a method that validates the dto
func (p *PriceCrud) Validate(repo port.Repository) error {
	if p.Action != "get" && p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *PriceCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns the domain of the dto
func (p *PriceCrud) GetDomain() []port.Domain {
	if p.Action == "add" && p.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		p.Date = time.Now().Format(pkg.DateFormat)
	}
	return []port.Domain{
		domain.NewPrice(p.ID, p.Date, p.Name, p.Unit, p.Pack),
	}
}

// GetOut is a method that returns the dto out
func (p *PriceCrud) GetOut() port.DTOOut {
	return &PriceCrud{}
}

// GetDTO is a method that returns the dto
func (p *PriceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	prices := slices[0].(*[]domain.Price)
	for _, domain := range *prices {
		unit := ""
		if domain.Unit != nil {
			unit = fmt.Sprintf("%.2f", *domain.Unit)
		}
		pack := ""
		if domain.Pack != nil {
			pack = fmt.Sprintf("%.2f", *domain.Pack)
		}
		ret = append(ret, &PriceCrud{
			ID:   domain.ID,
			Date: domain.Date.Format(pkg.DateFormat),
			Name: domain.Name,
			Unit: unit,
			Pack: pack,
		})
	}
	return ret
}

// isEmpty is a method that returns true if the dto is empty
func (p *PriceCrud) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.Name == "" && p.Unit == "" && p.Pack == ""
}
