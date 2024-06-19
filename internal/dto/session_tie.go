package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionTie represents the dto for tying a session
type SessionTie struct {
	Object string `json:"-" command:"name:session;key;pos:2-"`
	Action string `json:"-" command:"name:tie,untie;key;pos:2-"`
	ID     string `json:"id" command:"name:id;pos:3+"`
}

// SessionTieOut represents the dto for tying a session on output
type SessionTieOut struct {
	ID       string `json:"id" command:"name:id"`
	Process  string `json:"process" command:"name:process"`
	AgendaID string `json:"message" command:"name:agenda"`
}

// Validate is a method that validates the dto
func (s *SessionTie) Validate(repo port.Repository) error {
	if s.ID == "" {
		return errors.New(pkg.ErrIdUninformed)
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
		&domain.Session{
			ID: s.ID,
		},
	}
}

// GetOut is a method that returns the output dto
func (s *SessionTie) GetOut() port.DTOOut {
	return &SessionTieOut{}
}

// GetInstructions is a method that returns the instructions of the dto for a given domain
func (s *SessionTie) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return domain, []interface{}{s.ID}, nil
}

// GetDTO is a method that returns the dto
func (s *SessionTieOut) GetDTO(domainIn interface{}) []port.DTOOut {
	domain := domainIn.(*domain.Session)
	return []port.DTOOut{
		&SessionTieOut{
			ID:       domain.ID,
			Process:  domain.Process,
			AgendaID: domain.AgendaID,
		},
	}
}
