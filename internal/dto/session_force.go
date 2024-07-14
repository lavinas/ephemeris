package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionForce represents the dto for linking a session with agenda
type SessionForce struct {
	Base
	Object   string `json:"-" command:"name:session;key;pos:2-"`
	Action   string `json:"-" command:"name:force;key;pos:2-"`
	Sort     string `json:"sort" command:"name:sort;pos:3+"`
	Csv      string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID       string `json:"id" command:"name:id;pos:3+"`
	AgendaID string `json:"agenda" command:"name:agenda;pos:3+"`
}

// SessionForceOut represents the dto for linking a session with agenda on output
type SessionForceOut struct {
	Sort     string `json:"sort" command:"name:sort;pos:3+"`
	ID       string `json:"id" command:"name:id;pos:3+"`
	AgendaID string `json:"agenda" command:"name:agenda;pos:3+"`
	Process  string `json:"process" command:"name:process"`
}

// Validate is a method that validates the dto
func (s *SessionForce) Validate() error {
	if s.ID == "" && s.Csv == "" && s.AgendaID == "" {
		return errors.New(pkg.ErrInvalidParameters)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *SessionForce) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (s *SessionForce) GetDomain() []port.Domain {
	if s.Csv != "" {
		domains := []port.Domain{}
		sessions := []*SessionForce{}
		s.ReadCSV(&sessions, s.Csv)
		for _, session := range sessions {
			session.Action = s.Action
			session.Object = s.Object
			domains = append(domains, &domain.Session{ID: session.ID, AgendaID: session.AgendaID})
		}
		return domains
	}
	return []port.Domain{&domain.Session{ID: s.ID, AgendaID: s.AgendaID}}
}

// GetOut is a method that returns the output dto
func (s *SessionForce) GetOut() port.DTOOut {
	return &SessionForceOut{Sort: s.Sort, ID: s.ID}
}

// GetInstructions is a method that returns the instructions of the dto for a given domain
func (s *SessionForce) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return nil, nil, nil
}

// GetDto is a method that returns the dto
func (s *SessionForceOut) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		session := slice.(*domain.Session)
		ret = append(ret, &SessionForceOut{ID: session.ID, Process: session.Process, AgendaID: session.AgendaID})
	}
	pkg.NewCommands().Sort(ret, s.Sort)
	return ret
}
