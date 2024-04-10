package dto

import (
	"errors"
	"strconv"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
)

// RecurrenceGetIn is a struct that represents the recurrence get data transfer object
type RecurrenceGetIn struct {
	Object string `json:"-" command:"name:recurrence;key"`
	Action string `json:"-" command:"name:get;key"`
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Cycle  string `json:"cycle" command:"name:cycle"`
	Length string `json:"quantity" command:"name:len"`
	Limit  string `json:"limit" command:"name:limit"`
}

// RecurrenceGetOut is a struct that represents the recurrence get output data transfer object
type RecurrenceGetOut struct {
	ID     string `json:"id" command:"name:id"`
	Date   string `json:"date" command:"name:date"`
	Name   string `json:"name" command:"name:name"`
	Cycle  string `json:"cycle" command:"name:cycle"`
	Length string `json:"quantity" command:"name:len"`
	Limit  string `json:"limit" command:"name:limit"`
}

// Validate is a method that validates the dto
func (r *RecurrenceGetIn) Validate(repo port.Repository) error {
	if r.isEmpty() {
		return errors.New(port.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns the domain of the dto
func (r *RecurrenceGetIn) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewRecurrence(r.ID, r.Date, r.Name, r.Cycle, r.Length, r.Limit),
	}
}

// GetOut is a method that returns the dto out
func (r *RecurrenceGetIn) GetOut() port.DTOOut {
	return &RecurrenceGetOut{}
}

// GetDTO is a method that returns the dto
func (r *RecurrenceGetOut) GetDTO(domainIn interface{}) interface{} {
	ret := []RecurrenceGetOut{}
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
		ret = append(ret, RecurrenceGetOut{
			ID:     domain.ID,
			Date:   domain.Date.Format(port.DateFormat),
			Name:   domain.Name,
			Cycle:  domain.Cycle,
			Length: len,
			Limit:  lim,
		})
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (r *RecurrenceGetIn) isEmpty() bool {
	return r.ID == "" && r.Name == "" && r.Cycle == "" && 
	       r.Length == "" && r.Limit == "" && r.Date == ""
}
