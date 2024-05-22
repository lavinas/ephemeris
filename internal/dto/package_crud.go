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
	Object       string `json:"-" command:"name:package;key;pos:2-"`
	Action       string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort         string `json:"sort" command:"name:sort;pos:3+"`
	Csv          string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID           string `json:"id" command:"name:id;pos:3+;trans:id,string"`
	Date         string `json:"date" command:"name:date;pos:3+;trans:date,time"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence;pos:3+;trans:recurrence_id,string"`
	ServiceID    string `json:"service" command:"name:service;pos:3+;trans:service_id,string"`
	UnitValue    string `json:"unit" command:"name:unit;pos:3+;trans:price,numeric"`
	PackValue    string `json:"pack" command:"name:pack;pos:3+;trans:price,numeric"`
	Sequence     string `json:"seq" command:"name:seq;pos:3+;trans:sequence,numeric"`
	SequenceUp   string `json:"sequp" command:"name:sequp;pos:3+;trans:sequence,numeric"`
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
	itemId := ""
	if p.Action == "add" {
		if p.Date == "" {
			time.Local, _ = time.LoadLocation(pkg.Location)
			p.Date = time.Now().Format(pkg.DateFormat)
		}
		if p.UnitValue == "" {
			p.UnitValue = "0"
		}
		if p.PackValue == "" {
			p.PackValue = "0"
		}
		if p.Sequence == "" {
			p.Sequence = "0"
		}
		seq, _ := strconv.Atoi(p.Sequence)
		itemId = fmt.Sprintf("%s_%03d", p.ID, seq)
	}
	seqUp := p.Sequence
	if p.Action == "up" {
		if p.Sequence == "" {
			p.Sequence = "0"
		}
		seq, _ := strconv.Atoi(p.Sequence)
		itemId = fmt.Sprintf("%s_%03d", p.ID, seq)
		if p.SequenceUp != "" {
			seqUp = p.SequenceUp
		}
	}
	return []port.Domain{
		domain.NewPackage(p.ID, p.Date, p.ServiceID, p.RecurrenceID, p.PackValue),
		domain.NewPackageItem(itemId, p.ID, p.ServiceID, seqUp, p.UnitValue),
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
	packages := slices[0].(*[]domain.Package)
	items := slices[1].(*[]domain.PackageItem)
	packMap := make(map[string]*domain.Package)
	for _, p := range *packages {
		packMap[p.ID] = &p
	}
	last := ""
	for _, i := range *items {
		seq := "0"
		if i.Sequence != nil {
			seq = fmt.Sprintf("%d", *i.Sequence)
		}
		if pack, ok := packMap[i.PackageID]; ok {
			id := pack.ID
			date := pack.Date.Format(pkg.DateFormat)
			recurrence := pack.RecurrenceID
			if last == id {
				id = ""
				date = ""
				recurrence = ""
			}
			last = pack.ID
			ret = append(ret, &PackageCrud{
				ID:           id,
				Date:         date,
				RecurrenceID: recurrence,
				ServiceID:    i.ServiceID,
				UnitValue:    fmt.Sprintf("%.2f", *i.Price),
				PackValue:    fmt.Sprintf("%.2f", *pack.Price),
				Sequence:     seq,
			})
		}
	}
	pkg.NewCommands().Sort(ret, p.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (p *PackageCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	cmd, err := pkg.NewCommands().Transpose(p)
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) > 0 {
		d := p.GetDomain()
		domain := d[0]
		if fmt.Sprintf("%T", domain) != "*domain.Package" {
			domain = d[1]
		}
		return domain, cmd, nil
	}
	return domain, cmd, nil
}

// isEmpty is a method that checks if the dto is empty
func (p *PackageCrud) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.ServiceID == "" &&
		p.RecurrenceID == "" && p.UnitValue == "" && p.PackValue == "" && p.Sequence == ""
}
