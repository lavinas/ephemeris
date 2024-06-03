package dto

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// PackageAddIn represents the input dto for adding a package usecase
type PackageCrud struct {
	Base
	Object           string `json:"-" command:"name:package;key;pos:2-"`
	Action           string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort             string `json:"sort" command:"name:sort;pos:3+"`
	Csv              string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID               string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date             string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	RecurrenceID     string `json:"recurrence" command:"name:recurrence;pos:3+;trans:recurrence_id,string" csv:"recurrence"`
	ServiceID        string `json:"service" command:"name:service;pos:3+;trans:service_id,string" csv:"service"`
	UnitValue        string `json:"unit" command:"name:unit;pos:3+;trans:price,numeric" csv:"unit"`
	PackValue        string `json:"pack" command:"name:pack;pos:3+;trans:price,numeric" csv:"pack"`
	Sequence         string `json:"seq" command:"name:seq;pos:3+;trans:sequence,numeric" csv:"sequence"`
	SequenceUp       string `json:"sequp" command:"name:sequp;pos:3+;trans:sequence,numeric" csv:"sequp"`
	ItemInstructions []string
}

// Validate is a method that validates the dto
func (p *PackageCrud) Validate(repo port.Repository) error {
	if p.Csv != "" && (p.ID != "" || p.Date != "" || p.RecurrenceID != "" || p.ServiceID != "" ||
		p.UnitValue != "" || p.PackValue != "" || p.Sequence != "" || p.SequenceUp != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *PackageCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns a domain representation of the package dto
func (p *PackageCrud) GetDomain() []port.Domain {
	if p.Csv != "" {
		domains := []port.Domain{}
		packages := []*PackageCrud{}
		p.ReadCSV(&packages, p.Csv)
		for _, pack := range packages {
			pack.Action = p.Action
			pack.Object = p.Object
			domains = append(domains, p.getDomain(pack)...)
		}
		return domains
	}
	return p.getDomain(p)
}

// getDomain is a method that returns a domain representation of the package dto
func (x *PackageCrud) getDomain(one *PackageCrud) []port.Domain {
	itemId := ""
	if one.Action == "add" {
		if one.Date == "" {
			time.Local, _ = time.LoadLocation(pkg.Location)
			one.Date = time.Now().Format(pkg.DateFormat)
		}
		if one.UnitValue == "" {
			one.UnitValue = "0"
		}
		if one.PackValue == "" {
			one.PackValue = "0"
		}
		if one.Sequence == "" {
			one.Sequence = "0"
		}
		seq, _ := strconv.Atoi(one.Sequence)
		itemId = fmt.Sprintf("%s_%03d", one.ID, seq)
	}
	seqUp := one.Sequence
	if one.Action == "up" {
		if one.Sequence == "" {
			one.Sequence = "0"
		}
		seq, _ := strconv.Atoi(one.Sequence)
		itemId = fmt.Sprintf("%s_%03d", one.ID, seq)
		if one.SequenceUp != "" {
			seqUp = one.SequenceUp
		}
	}
	return []port.Domain{
		domain.NewPackage(one.ID, one.Date, one.RecurrenceID, one.PackValue),
		domain.NewPackageItem(itemId, one.ID, one.ServiceID, seqUp, one.UnitValue),
	}
}

// GetOut is a method that returns the output dto
func (p *PackageCrud) GetOut() port.DTOOut {
	return p
}

// GetDTO is a method that returns the dto
func (p *PackageCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for i := 0; i < len(slices); i += 2 {
		pack := slices[i].(*[]domain.Package)
		packMap := make(map[string]*domain.Package)
		for _, p := range *pack {
			packMap[p.ID] = &p
		}
		item := slices[i+1].(*[]domain.PackageItem)
		for _, i := range *item {
			if pack, ok := packMap[i.PackageID]; ok {
				ret = append(ret, &PackageCrud{
					ID:           pack.ID,
					Date:         pack.Date.Format(pkg.DateFormat),
					RecurrenceID: pack.RecurrenceID,
					ServiceID:    i.ServiceID,
					UnitValue:    fmt.Sprintf("%.2f", *i.Price),
					PackValue:    fmt.Sprintf("%.2f", *pack.Price),
					Sequence:     fmt.Sprintf("%d", *i.Sequence),
				})
			}
		}
	}
	pkg.NewCommands().Sort(ret, p.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (p *PackageCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return domain, []interface{}{}, nil
}
