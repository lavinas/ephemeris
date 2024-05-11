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
	Object    string `json:"-" command:"name:package;key;pos:2-"`
	Action    string `json:"-" command:"name:append;key;pos:2-"`
	Sort      string `json:"sort" command:"name:sort;pos:3+"`
	ID        string `json:"id" command:"name:id;pos:3+"`
	ServiceID string `json:"service" command:"name:service;pos:3+"`
	UnitValue string `json:"unit" command:"name:unit;pos:3+"`
	Sequence  string `json:"seq" command:"name:seq;pos:3+"`
}

// Validate is a method that validates the dto
func (p *PackageAppend) Validate(repo port.Repository) error {
	if p.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *PackageAppend) GetCommand() string {
	return "add"
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageAppend) GetDomain() []port.Domain {
	if p.UnitValue == "" {
		p.UnitValue = "0"
	}
	seq, _ := strconv.Atoi(p.Sequence)
	itemId := fmt.Sprintf("%s_%03d", p.ID, seq)
	return []port.Domain{
		domain.NewPackageItem(itemId, p.ID, p.ServiceID, p.Sequence, p.UnitValue),
	}
}

// GetOut is a method that returns the dto out
func (p *PackageAppend) GetOut() port.DTOOut {
	return p
}

// GetDTO is a method that returns the dto
func (p *PackageAppend) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	items := slices[0].(*[]domain.PackageItem)
	for _, domain := range *items {
		unit := ""
		if domain.Price != nil {
			unit = fmt.Sprintf("%.2f", *domain.Price)
		}
		ret = append(ret, &PackageAppend{
			ID:        domain.PackageID,
			ServiceID: domain.ServiceID,
			UnitValue: unit,
			Sequence:  fmt.Sprintf("%d", *domain.Sequence),
		})
	}
	pkg.NewCommands().Sort(ret, p.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (p *PackageAppend) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	cmd, err := pkg.NewCommands().Transpose(p)
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) > 0 {
		domain := p.GetDomain()[0]
		return domain, cmd, nil
	}
	return domain, cmd, nil
}

// isEmpty is a method that checks if the dto is empty
func (p *PackageAppend) isEmpty() bool {
	return p.ID == "" && p.ServiceID == "" && p.UnitValue == "" && p.Sequence == ""
}
