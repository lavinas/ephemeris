package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// RecurrenceAddIn is a struct that represents the recurrence add data transfer object
type RecurrenceAddIn struct {
	Object string `json:"-" command:"name:recurrence;key"`
	Action string `json:"-" command:"name:add;key"`
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Cycle  string `json:"cycle" command:"name:cycle"`
	Length string `json:"quantity" command:"name:length"`
	Limit  string `json:"limit" command:"name:limit"`
}

// RecurrenceAddOut is a struct that represents the recurrence add output data transfer object
type RecurrenceAddOut struct {
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Cycle  string `json:"cycle" command:"name:cycle"`
	Length string `json:"quantity" command:"name:length"`
	Limit  string `json:"limit" command:"name:limit"`
}

// Validate is a method that validates the dto
func (r *RecurrenceAddIn) Validate(repo port.Repository) error {
	if r.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (r *RecurrenceAddIn) GetDomain() []port.Domain {
	if r.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		r.Date = time.Now().Format(pkg.DateFormat)
	}
	if r.Length == "" {
		r.Length = "0"
	}
	if r.Limit == "" {
		r.Limit = "0"
	}
	return []port.Domain{
		domain.NewRecurrence(r.ID, r.Date, r.Name, r.Cycle, r.Length, r.Limit),
	}
}

// GetOut is a method that returns the dto out
func (r *RecurrenceAddIn) GetOut() port.DTOOut {
	return &RecurrenceAddOut{}
}

// GetDTO is a method that returns the dto
func (r *RecurrenceAddOut) GetDTO(domainIn interface{}) interface{} {
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
	return &RecurrenceAddOut{
		ID:     recurrence.ID,
		Date:   recurrence.Date.Format(pkg.DateFormat),
		Name:   recurrence.Name,
		Cycle:  recurrence.Cycle,
		Length: len,
		Limit:  lim,
	}
}

// isEmpty is a method that checks if the dto is empty
func (r *RecurrenceAddIn) isEmpty() bool {
	return r.ID == "" && r.Date == "" && r.Name == "" && r.Cycle == "" && r.Length == "" && r.Limit == ""
}
