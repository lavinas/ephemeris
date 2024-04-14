package dto

import (
	"errors"
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PriceGetIn is a struct that represents the price get data transfer object
type PriceGetIn struct {
	Object string `json:"-" command:"name:price;key;pos:2-"`
	Action string `json:"-" command:"name:get;key;pos:2-"`
	ID     string `json:"id" command:"name:id;pos:3+"`
	Date   string `json:"date" command:"name:date;pos:3+"`
	Name   string `json:"name" command:"name:name;pos:3+"`
	Unit   string `json:"unit" command:"name:unit;pos:3+"`
	Pack   string `json:"pack" command:"name:pack;pos:3+"`
}

// PriceGetOut is a struct that represents the price get output data transfer object
type PriceGetOut struct {
	ID   string `json:"id" command:"name:id"`
	Date string `json:"date" command:"name:date"`
	Name string `json:"name" command:"name:name"`
	Unit string `json:"unit" command:"name:unit"`
	Pack string `json:"pack" command:"name:pack"`
}

// Validate is a method that validates the dto
func (p *PriceGetIn) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (p *PriceGetIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewPrice(p.ID, p.Date, p.Name, p.Unit, p.Pack),
	}
}

// GetOut is a method that returns the dto out
func (p *PriceGetIn) GetOut() port.DTOOut {
	return &PriceGetOut{}
}

// GetDTO is a method that returns the dto
func (p *PriceGetOut) GetDTO(domainIn interface{}) []port.DTOOut {
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
		ret = append(ret, &PriceGetOut{
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
func (p *PriceGetIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.Name == "" && p.Unit == "" && p.Pack == ""
}
