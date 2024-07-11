package dto

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PackageAppend is a struct that represents the package append dto
type PackageAppend struct {
	Base
	Object    string `json:"-" command:"name:package;key;pos:2-"`
	Action    string `json:"-" command:"name:append;key;pos:2-"`
	Sort      string `json:"sort" command:"name:sort;pos:3+"`
	Csv       string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID        string `json:"id" command:"name:id;pos:3+" csv:"id"`
	ServiceID string `json:"service" command:"name:service;pos:3+" csv:"service"`
	UnitValue string `json:"unit" command:"name:unit;pos:3+" csv:"unit"`
	Sequence  string `json:"seq" command:"name:seq;pos:3+" csv:"seq"`
}

// Validate is a method that validates the dto
func (p *PackageAppend) Validate() error {
	if p.Csv != "" && (p.ID != "" || p.ServiceID != "" || p.UnitValue != "" || p.Sequence != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *PackageAppend) GetCommand() string {
	return "add"
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageAppend) GetDomain() []port.Domain {
	if p.Csv != "" {
		domains := []port.Domain{}
		items := []*PackageAppend{}
		p.ReadCSV(&items, p.Csv)
		for _, item := range items {
			item.Action = p.Action
			item.Object = p.Object
			domains = append(domains, p.getDomain(item))
		}
		return domains
	}
	return []port.Domain{p.getDomain(p)}
}

// getDomain is a method that returns a domain representation of the package dto
func (p *PackageAppend) getDomain(one *PackageAppend) port.Domain {
	if one.Action == "add" && one.UnitValue == "" {
		one.UnitValue = "0"
	}
	seq, _ := strconv.Atoi(one.Sequence)
	itemId := fmt.Sprintf("%s_%03d", one.ID, seq)
	return domain.NewPackageItem(itemId, one.ID, one.ServiceID, one.Sequence, one.UnitValue)
}

// GetOut is a method that returns the dto out
func (p *PackageAppend) GetOut() port.DTOOut {
	return p
}

// GetDTO is a method that returns the dto
func (p *PackageAppend) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		items := slice.(*[]domain.PackageItem)
		for _, item := range *items {
			unit := ""
			if item.Price != nil {
				unit = fmt.Sprintf("%.2f", *item.Price)
			}
			ret = append(ret, &PackageAppend{
				ID:        item.PackageID,
				ServiceID: item.ServiceID,
				UnitValue: unit,
				Sequence:  fmt.Sprintf("%d", *item.Sequence),
			})
		}
	}
	pkg.NewCommands().Sort(ret, p.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (p *PackageAppend) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return p.getInstructions(p, domain)
}
