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
	Base
	Object string `json:"-" command:"name:recurrence;key;pos:2-"`
	Action string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort   string `json:"sort" command:"name:sort;pos:3+"`
	Csv    string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID     string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date   string `json:"date" command:"name:date;pos:3+;trans:date,time"  csv:"date"`
	Name   string `json:"name" command:"name:name;pos:3+;trans:name,string" csv:"name"`
	Cycle  string `json:"cycle" command:"name:cycle;pos:3+;trans:cycle,string" csv:"cycle"`
	Length string `json:"quantity" command:"name:length;pos:3+;trans:length,numeric" csv:"length"`
	Limit  string `json:"limit" command:"name:limit;pos:3+;trans:limit,numeric" csv:"limit"`
}

// Validate is a method that validates the dto
func (r *RecurrenceCrud) Validate() error {
	if r.Csv != "" && (r.ID != "" || r.Date != "" || r.Cycle != "" || r.Length != "" || r.Limit != "" || r.Name != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (p *RecurrenceCrud) GetCommand() string {
	return p.Action
}

// GetDomain is a method that returns the domain of the dto
func (r *RecurrenceCrud) GetDomain() []port.Domain {
	if r.Csv != "" {
		domains := []port.Domain{}
		recurrences := []*RecurrenceCrud{}
		r.ReadCSV(&recurrences, r.Csv)
		for _, recurrence := range recurrences {
			recurrence.Action = r.Action
			recurrence.Object = r.Object
			domains = append(domains, r.getDomain(recurrence))
		}
		return domains
	}
	return []port.Domain{r.getDomain(r)}
}

// getDomain is a method that returns the domain of the dto
func (r *RecurrenceCrud) getDomain(one *RecurrenceCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Length == "" {
		one.Length = "0"
	}
	if one.Action == "add" && one.Limit == "" {
		one.Limit = "0"
	}
	if one.Action == "add" && one.Cycle == "" {
		one.Cycle = pkg.DefaultRecurrenceCycle
	}
	return domain.NewRecurrence(one.ID, one.Date, one.Name, one.Cycle, one.Length, one.Limit)
}

// GetOut is a method that returns the dto out
func (r *RecurrenceCrud) GetOut() port.DTOOut {
	return r
}

// GetDTO is a method that returns the dto
func (r *RecurrenceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		recurrences := slice.(*[]domain.Recurrence)
		for _, recurrence := range *recurrences {
			len := ""
			if recurrence.Length != nil {
				len = strconv.FormatInt(*recurrence.Length, 10)
			}
			lim := ""
			if recurrence.Limits != nil {
				lim = strconv.FormatInt(*recurrence.Limits, 10)
			}
			ret = append(ret, &RecurrenceCrud{
				ID:     recurrence.ID,
				Date:   recurrence.Date.Format(pkg.DateFormat),
				Name:   recurrence.Name,
				Cycle:  recurrence.Cycle,
				Length: len,
				Limit:  lim,
			})
		}
	}
	pkg.NewCommands().Sort(ret, r.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (r *RecurrenceCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return r.getInstructions(r, domain)
}
