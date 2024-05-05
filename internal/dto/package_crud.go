package dto

import (
	"errors"
	"fmt"
	"time"
	"strconv"

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
	Sequence     string `json:"seq" command:"name:seq;pos:3+"`
	SequenceUp   string `json:"sequp" command:"name:sequp;pos:3+"`
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
	return &PackageCrud{}
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
	for _, i := range *items {
		seq := "0"
		if i.Sequence != nil {
			seq = fmt.Sprintf("%d", *i.Sequence)
		}
		if pack, ok := packMap[i.PackageID]; ok {
			ret = append(ret, &PackageCrud{
				ID:           pack.ID,
				Date:         pack.Date.Format(pkg.DateFormat),
				ServiceID:    i.ServiceID,
				RecurrenceID: pack.RecurrenceID,
				UnitValue:    fmt.Sprintf("%.2f", *i.Price),
				PackValue:    fmt.Sprintf("%.2f", *pack.Price),
				Sequence:     seq,
			})
		}
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (p *PackageCrud) isEmpty() bool {
	return p.ID == "" && p.Date == "" && p.ServiceID == "" &&
		p.RecurrenceID == "" && p.UnitValue == "" && p.PackValue == "" && p.Sequence == ""
}
