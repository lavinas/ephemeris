package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionTie represents the dto for tying a session
type SessionTie struct {
	Base
	Object    string `json:"-" command:"name:session;key;pos:2-"`
	Action    string `json:"-" command:"name:tie,untie;key;pos:2-"`
	Sort      string `json:"sort" command:"name:sort;pos:3+"`
	ID        string `json:"id" command:"name:id;pos:3+"`
	ClientID  string `json:"client" command:"name:client;pos:3+;trans:client_id,string"`
	ServiceID string `json:"service" command:"name:service;pos:3+;trans:service_id,string"`
	At        string `json:"at" command:"name:at;pos:3+;trans:at,time"`
	Status    string `json:"status" command:"name:status;pos:3+;trans:status,string"`
	Process   string `json:"process" command:"name:process;pos:3;trans:process,string"`
}

// SessionTieOut represents the dto for tying a session on output
type SessionTieOut struct {
	Sort      string `json:"sort" command:"name:sort;pos:3+"`
	ID        string `json:"id" command:"name:id"`
	ClientID  string `json:"client" command:"name:client"`
	ServiceID string `json:"service" command:"name:service"`
	At        string `json:"at" command:"name:at"`
	Status    string `json:"status" command:"name:status"`
	Process   string `json:"process" command:"name:process"`
	AgendaID  string `json:"message" command:"name:agenda/error"`
}

// Validate is a method that validates the dto
func (s *SessionTie) Validate(repo port.Repository) error {
	if s.ID == "" && s.ClientID == "" && s.ServiceID == "" && s.At == "" &&
		s.Status == "" && s.Process == "" {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	if err := s.validateAt(); err != nil {
		return err
	}
	return nil
}

// validateAt is a method that validates the at field
func (s *SessionTie) validateAt() error {
	sTest := &SessionTie{At: s.At}
	_, err := pkg.NewCommands().Transpose(sTest)
	if err != nil {
		return errors.New(pkg.ErrInvalidAt)
	}
	if sTest.At == "" {
		return nil
	}
	if _, err = time.Parse(pkg.DateTimeFormat, sTest.At); err != nil {
		if _, err = time.Parse(pkg.DateFormat, sTest.At); err != nil {
			return errors.New(pkg.ErrInvalidAt)
		}
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *SessionTie) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (s *SessionTie) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewSession(s.ID, "", "", s.ClientID, s.ServiceID, s.At, s.Status, s.Process, ""),
	}
}

// GetOut is a method that returns the output dto
func (s *SessionTie) GetOut() port.DTOOut {
	return &SessionTieOut{Sort: s.Sort}
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (s *SessionTie) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return s.getInstructions(s, domain)
}

// GetDTO is a method that returns the dto
func (s *SessionTieOut) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		domain := slice.(*domain.Session)
		ret = append(ret, &SessionTieOut{
			ID:        domain.ID,
			ClientID:  domain.ClientID,
			ServiceID: domain.ServiceID,
			At:        domain.At.Format(pkg.DateTimeFormat),
			Status:    domain.Status,
			Process:   domain.Process,
			AgendaID:  domain.AgendaID,
		})
	}
	pkg.NewCommands().Sort(ret, s.Sort)
	return ret
}
