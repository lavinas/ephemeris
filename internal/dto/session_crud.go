package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionCrud represents the dto for getting a session
type SessionCrud struct {
	Base
	Object     string `json:"-" command:"name:session;key;pos:2-"`
	Action     string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort       string `json:"sort" command:"name:sort;pos:3+"`
	Csv	       string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID         string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date       string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	ClientID   string `json:"client" command:"name:client;pos:3+;trans:client_id,string" csv:"client"`
	ServiceID  string `json:"service" command:"name:service;pos:3+;trans:service_id,string" csv:"service"`
	At         string `json:"at" command:"name:at;pos:3+;trans:at,time" csv:"at"`
	Kind       string `json:"kind" command:"name:kind;pos:3+;trans:kind,string" csv:"kind"`
	Status     string `json:"status" command:"name:status;pos:3+;trans:status,string" csv:"status"`
}

// Validate is a method that validates the dto
func (s *SessionCrud) Validate(repo port.Repository) error {
	if s.Csv != "" && s.Action == "get" {
		return errors.New(pkg.ErrCsvAndGet)
	}
	if s.Csv != "" && (s.ID != "" || s.Date != "" || s.ClientID != "" || s.ServiceID != "" || s.At != "" || s.Kind != "" || s.Status != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *SessionCrud) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (s *SessionCrud) GetDomain() []port.Domain {
	domains := []port.Domain{}
	if s.Csv != "" {
		sessions := []*SessionCrud{}
		s.ReadCSV(&sessions, s.Csv)
		for _, se := range sessions {
			se.Action = s.Action
			se.Object = s.Object
			domains = append(domains, se.getDomain(se))
		}
		return domains
	}
	return []port.Domain{
		s.getDomain(s),
	}
}

// getDomain is a method that returns the domain of one object
func (s *SessionCrud) getDomain(one *SessionCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Kind == "" {
		one.Kind = pkg.DefaultSessionKind
	}
	if one.Action == "add" && one.Status == "" {
		one.Status = pkg.DefaultSessionStatus
	}
	if one.Action == "add" && one.ID == "" {
		at := time.Now().Format("2006-01-02-15-04")
		t, err := time.Parse(pkg.DateTimeFormat, one.At)
		if err != nil {
			t, err = time.Parse(pkg.DateFormat, one.At)
		}
		if err == nil {
			at = t.Format("2006-01-02-15-04")
		}
		one.ID = at + "-" + one.ClientID + "-" + one.ServiceID
	}
	return domain.NewSession(one.ID, one.Date, one.ClientID, one.ServiceID, one.At, one.Kind, one.Status)
}

// GetOut is a method that returns the output dto
func (s *SessionCrud) GetOut() port.DTOOut {
	return s
}

// GetDTO is a method that returns the dto
func (s *SessionCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		sessions := slice.(*[]domain.Session)
		for _, se := range *sessions {
			ret = append(ret, &SessionCrud{
				ID:         se.ID,
				Date:       se.Date.Format(pkg.DateFormat),
				ClientID:   se.ClientID,
				ServiceID:  se.ServiceID,
				At:         se.At.Format(pkg.DateTimeFormat),
				Kind:       se.Kind,
				Status:     se.Status,
			})
		}
	}
	pkg.NewCommands().Sort(ret, s.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (s *SessionCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return s.getInstructions(s, domain)
}	