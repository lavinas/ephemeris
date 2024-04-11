package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PriceAddIn is a struct that represents the price add data transfer object
type PriceAddIn struct {
	Object string `json:"-" command:"name:price;key"`
	Action string `json:"-" command:"name:add;key"`
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Unit   string `json:"unit" command:"name:unit"`
	Pack   string `json:"pack" command:"name:pack"`
}

// PriceAddOut is a struct that represents the price add output data transfer object
type PriceAddOut struct {
	ID   string `json:"id" command:"name:id"`
	Date string `json:"date" command:"name:date"`
	Name string `json:"name" command:"name:name"`
	Unit string `json:"unit" command:"name:unit"`
	Pack string `json:"pack" command:"name:pack"`
}

// Validate is a method that validates the dto
func (p *PriceAddIn) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (p *PriceAddIn) GetDomain() []port.Domain {
	if p.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		p.Date = time.Now().Format(pkg.DateFormat)
	}
	return []port.Domain{
		domain.NewPrice(p.ID, p.Date, p.Name, p.Unit, p.Pack),
	}
}

// GetOut is a method that returns the dto out
func (p *PriceAddIn) GetOut() port.DTOOut {
	return &PriceAddOut{}
}

// GetDTO is a method that returns the dto
func (p *PriceAddOut) GetDTO(domainIn interface{}) interface{} {
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
	return &PriceAddOut{
		ID:   price.ID,
		Date: price.Date.Format(pkg.DateFormat),
		Name: price.Name,
		Unit: unit,
		Pack: pack,
	}
}

// isEmpty is a method that checks if the dto is empty
func (p *PriceAddIn) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.Name == "" && p.Unit == "" && p.Pack == ""
}
