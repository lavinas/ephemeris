package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// RecurrenceCrud is a struct that represents the recurrence get data transfer object
type RecurrenceCrud struct {
	Object string `json:"-" command:"name:recurrence;key;pos:2-"`
	Action string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID     string `json:"id" command:"name:id;pos:3+"`
	Date   string `json:"date" command:"name:date;pos:3+"`
	Name   string `json:"name" command:"name:name;pos:3+"`
	Cycle  string `json:"cycle" command:"name:cycle;pos:3+"`
	Length string `json:"quantity" command:"name:length;pos:3+"`
	Limit  string `json:"limit" command:"name:limit;pos:3+"`
}

// Validate is a method that validates the dto
func (r *RecurrenceCrud) Validate(repo port.Repository) error {
	if r.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *RecurrenceCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns the domain of the dto
func (r *RecurrenceCrud) GetDomain() []port.Domain {
	if r.Action == "add" && r.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		r.Date = time.Now().Format(pkg.DateFormat)
	}
	if r.Action == "add" && r.Length == "" {
		r.Length = "0"
	}
	if r.Action == "add" && r.Limit == "" {
		r.Limit = "0"
	}
	if r.Action == "add" && r.Cycle == "" {
		r.Cycle = pkg.DefaultRecurrenceCycle
	}
	return []port.Domain{
		domain.NewRecurrence(r.ID, r.Date, r.Name, r.Cycle, r.Length, r.Limit),
	}
}

// GetOut is a method that returns the dto out
func (r *RecurrenceCrud) GetOut() port.DTOOut {
	return &RecurrenceCrud{}
}

// GetDTO is a method that returns the dto
func (r *RecurrenceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	recurrences := slices[0].(*[]domain.Recurrence)
	for _, domain := range *recurrences {
		len := ""
		if domain.Length != nil {
			len = strconv.FormatInt(*domain.Length, 10)
		}
		lim := ""
		if domain.Limits != nil {
			lim = strconv.FormatInt(*domain.Limits, 10)
		}
		ret = append(ret, &RecurrenceCrud{
			ID:     domain.ID,
			Date:   domain.Date.Format(pkg.DateFormat),
			Name:   domain.Name,
			Cycle:  domain.Cycle,
			Length: len,
			Limit:  lim,
		})
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (r *RecurrenceCrud) isEmpty() bool {
	return r.ID == "" && r.Name == "" && r.Cycle == "" &&
		r.Length == "" && r.Limit == "" && r.Date == ""
}