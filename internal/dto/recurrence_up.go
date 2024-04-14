package dto

import (
	"errors"
	"strconv"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// RecurrenceUpIn is a struct that represents the recurrence up data transfer object
type RecurrenceUpIn struct {
	Object string `json:"-" command:"name:recurrence;key;pos:2-"`
	Action string `json:"-" command:"name:up;key;pos:2-"`
	ID     string `json:"id" command:"name:id;pos:3+"`
	Date   string `json:"date" command:"name:date;pos:3+"`
	Name   string `json:"name" command:"name:name;pos:3+"`
	Cycle  string `json:"cycle" command:"name:cycle;pos:3+"`
	Length string `json:"quantity" command:"name:length;pos:3+"`
	Limit  string `json:"limit" command:"name:limit;pos:3+"`
}

// RecurrenceUpOut is a struct that represents the recurrence up output data transfer object
type RecurrenceUpOut struct {
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Cycle  string `json:"cycle" command:"name:cycle"`
	Length string `json:"quantity" command:"name:length"`
	Limit  string `json:"limit" command:"name:limit"`
}

// Validate is a method that validates the dto
func (r *RecurrenceUpIn) Validate(repo port.Repository) error {
	if r.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	if r.ID == "" {
		return errors.New(pkg.ErrIdUninformed)
	}
	id := r.ID
	r.ID = ""
	if r.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	r.ID = id
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (r *RecurrenceUpIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewRecurrence(r.ID, r.Date, r.Name, r.Cycle, r.Length, r.Limit),
	}
}

// GetOut is a method that returns the dto out
func (r *RecurrenceUpIn) GetOut() port.DTOOut {
	return &RecurrenceUpOut{}
}

// GetDTO is a method that returns the dto
func (r *RecurrenceUpOut) GetDTO(domainIn interface{}) []port.DTOOut {
	slices := domainIn.([]interface{})
	recurrence, ok := slices[0].(*domain.Recurrence)
	if !ok {
		return nil
	}
	len := ""
	if recurrence.Length != nil {
		len = strconv.FormatInt(*recurrence.Length, 10)
	}
	lim := ""
	if recurrence.Limits != nil {
		lim = strconv.FormatInt(*recurrence.Limits, 10)
	}
	return []port.DTOOut{&RecurrenceUpOut{
		ID:     recurrence.ID,
		Date:   recurrence.Date.Format(pkg.DateFormat),
		Name:   recurrence.Name,
		Cycle:  recurrence.Cycle,
		Length: len,
		Limit:  lim,
	}}
}

// isEmpty is a method that checks if the dto is empty
func (r *RecurrenceUpIn) isEmpty() bool {
	return r.ID == "" && r.Date == "" && r.Name == "" && r.Cycle == "" && r.Length == "" && r.Limit == ""
}
