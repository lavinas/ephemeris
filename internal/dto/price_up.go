package dto

import (
	"errors"
	"fmt"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PriceUpIn is a struct that represents the price up data transfer object
type PriceUpIn struct {
	Object string `json:"-" command:"name:price;key"`
	Action string `json:"-" command:"name:up;key"`
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Unit   string `json:"unit" command:"name:unit"`
	Pack   string `json:"pack" command:"name:pack"`
}

// PriceUpOut is a struct that represents the price up output data transfer object
type PriceUpOut struct {
	ID   string `json:"id" command:"name:id"`
	Date string `json:"date" command:"name:date"`
	Name string `json:"name" command:"name:name"`
	Unit string `json:"unit" command:"name:unit"`
	Pack string `json:"pack" command:"name:pack"`
}

// Validate is a method that validates the dto
func (p *PriceUpIn) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	if p.ID == "" {
		return errors.New(pkg.ErrIdUninformed)
	}
	id := p.ID
	p.ID = ""
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	p.ID = id
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (p *PriceUpIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewPrice(p.ID, p.Date, p.Name, p.Unit, p.Pack),
	}
}

// GetOut is a method that returns the dto out
func (p *PriceUpIn) GetOut() port.DTOOut {
	return &PriceUpOut{}
}

// GetDTO is a method that returns the dto
func (p *PriceUpOut) GetDTO(domainIn interface{}) interface{} {
	slices := domainIn.([]interface{})
	price := slices[0].(*domain.Price)
	unit := ""
	if price.Unit != nil {
		unit = fmt.Sprintf("%.2f", *price.Unit)
	}
	pack := ""
	if price.Pack != nil {
		pack = fmt.Sprintf("%.2f", *price.Pack)
	}
	return &PriceUpOut{
		ID:   price.ID,
		Date: price.Date.Format(pkg.DateFormat),
		Name: price.Name,
		Unit: unit,
		Pack: pack,
	}
}

// isEmpty is a method that checks if the dto is empty
func (p *PriceUpIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.Name == "" && p.Unit == "" && p.Pack == ""
}
