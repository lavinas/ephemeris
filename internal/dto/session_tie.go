package dto

import (
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/internal/domain"
)

// SessionTie represents the dto for tying a session
type SessionTie struct {
	Object    string `json:"-" command:"name:session;key;pos:2-"`
	Action    string `json:"-" command:"name:tie;key;pos:2-"`
	ID        string `json:"id" command:"name:id;pos:3+"`    
}

// SessionTieOut represents the dto for tying a session on output
type SessionTieOut struct {
	ID        string `json:"id" command:"name:id"`
	Process   string `json:"process" command:"name:process"`
	Message   string `json:"message" command:"name:message"`
}


// Validate is a method that validates the dto
func (s *SessionTie) Validate(repo port.Repository) error {
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *SessionTie) GetCommand() string {
	return s.Action
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
			ID: domain.ID,
			Process: "ok",
			Message: "session tied",
		},
	}
}


