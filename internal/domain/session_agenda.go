package domain

import (
	"errors"
	"slices"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	statusSessionAgenda = []string{pkg.SessionAgendaStatusLinked,
		pkg.SessionAgendaStatusConfirm,
		pkg.SessionAgendaStatusConfirmed,
		pkg.SessionAgendaStatusCancelled}
)

// SessionAgenda represents the domain for a session linked to agenda
type SessionAgenda struct {
	ID        string `gorm:"type:varchar(301); primaryKey"`
	SessionID string `gorm:"type:varchar(150); not null; index"`
	AgendaID  string `gorm:"type:varchar(150); not null; index"`
	StatusID  string `gorm:"type:varchar(50); not null; index"`
}

// NewSessionAgenda creates a new session agenda domain entity
func NewSessionAgenda(sessionID, agendaID, status string) *SessionAgenda {
	sessionAgenda := &SessionAgenda{}
	sessionAgenda.SessionID = sessionID
	sessionAgenda.AgendaID = agendaID
	sessionAgenda.StatusID = status
	sessionAgenda.ID = sessionID + "_" + agendaID
	return sessionAgenda
}

// Validate validates the session agenda domain entity
func (s *SessionAgenda) Format(repo port.Repository, args ...string) error {
	message := ""
	filled := slices.Contains(args, "filled")
	if err := s.formatSessionID(repo, filled); err != nil {
		message += err.Error() + " | "
	}
	if err := s.formatAgendaID(repo, filled); err != nil {
		message += err.Error() + " | "
	}
	if err := s.formatStatusID(filled); err != nil {
		message += err.Error() + " | "
	}
	if err := s.validateDuplicity(repo, slices.Contains(args, "noduplicity")); err != nil {
		message += err.Error() + " | "
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

// validateSessionID validates the session id
func (s *SessionAgenda) formatSessionID(repo port.Repository, filled bool) error {
	if s.SessionID == "" {
		if filled {
			return nil
		}
		return errors.New("session id is required")
	}
	session := Session{ID: s.SessionID}
	if ok, err := session.Load(repo); err != nil {
		return err
	} else if !ok {
		return errors.New("session not found")
	}
	return nil
}

// validateAgendaID validates the agenda id
func (s *SessionAgenda) formatAgendaID(repo port.Repository, filled bool) error {
	if s.AgendaID == "" {
		if filled {
			return nil
		}
		return errors.New("agenda id is required")
	}
	agenda := Agenda{ID: s.AgendaID}
	if ok, err := agenda.Load(repo); err != nil {
		return err
	} else if !ok {
		return errors.New("agenda not found")
	}
	return nil
}

// validateStatusID validates the status id
func (s *SessionAgenda) formatStatusID(filled bool) error {
	if s.StatusID == "" {
		if filled {
			return nil
		}
		return errors.New("status id is required")
	}
	if !slices.Contains(statusSessionAgenda, s.StatusID) {
		return errors.New("status id is invalid")
	}
	return nil
}

// validateDuplicity validates the duplicity of the session agenda
func (s *SessionAgenda) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		if ok, err := repo.Get(s, s.ID); err != nil {
			return err
		} else if ok {
			return errors.New("session agenda already exists")
		}
	}
	return nil
}
