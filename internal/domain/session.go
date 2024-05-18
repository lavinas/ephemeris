package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	KindSession   = []string{pkg.SessionKindRegular, pkg.SessionKindRescheduled, pkg.SessionKindExtra}
	StatusSession = []string{pkg.SessionStatusOpen, pkg.SessionStatusDone, pkg.SessionStatusCanceled}
)

// Session represents the session entity
type Session struct {
	ID         string    `gorm:"type:varchar(50); primaryKey"`
	Date       time.Time `gorm:"type:datetime; not null"`
	ClientID   string    `gorm:"type:varchar(50); not null; index"`
	ServiceID  string    `gorm:"type:varchar(50); not null; index"`
	At         time.Time `gorm:"type:datetime; not null"`
	Kind       string    `gorm:"type:varchar(50); not null; index"`
	Status     string    `gorm:"type:varchar(50); not null; index"`
}

// NewSession creates a new session domain entity
func NewSession(id, date, clientID, serviceID, at, kind, status string) *Session {
	session := &Session{}
	session.ID = id
	session.ClientID = clientID
	session.ServiceID = serviceID
	local, _ := time.LoadLocation(pkg.Location)
	session.Date, _ = time.ParseInLocation(pkg.DateFormat, date, local)
	var err error
	session.At, err = time.ParseInLocation(pkg.DateTimeFormat, at, local)
	if err != nil {
		session.At, _ = time.ParseInLocation(pkg.DateFormat, at, local)
	}
	session.Kind = kind
	session.Status = status
	return session
}

// Validate is a method that validates the session entity
func (s *Session) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	msg := ""
	if err := s.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatClientID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatAt(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatKind(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.validateDuplicity(repo, slices.Contains(args, "noduplicity")); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a function that checks if a agenda exists
func (s *Session) Load(repo port.Repository) (bool, error) {
	return repo.Get(s, s.ID)
}

// GetID is a method that returns the id of the client
func (s *Session) GetID() string {
	return s.ID
}

// Get is a method that returns the client
func (s *Session) Get() port.Domain {
	return s
}

// GetEmpty is a method that returns an empty client with just id
func (s *Session) GetEmpty() port.Domain {
	return &Session{}
}

// TableName returns the table name for database
func (s *Session) TableName() string {
	return "session"
}

// formatID is a method that validates the session id
func (s *Session) formatID(filled bool) error {
	if s.ID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	if len(s.ID) > 50 {
		return errors.New(pkg.ErrLongID)
	}
	ids := strings.Split(s.ID, " ")
	if len(ids) > 1 {
		return errors.New(pkg.ErrInvalidID)
	}
	return nil
}

// formatDate is a method that validates the session date
func (s *Session) formatDate(filled bool) error {
	if s.Date.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
	}
	return nil
}

// validateClientID is a method that validates the session client id
func (s *Session) formatClientID(repo port.Repository, filled bool) error {
	if s.ClientID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyClientID)
	}
	client := &Client{ID: s.ClientID}
	if ok, err := client.Load(repo); err != nil {
		return err
	} else if !ok {
		return errors.New(pkg.ErrClientNotFound)
	}
	return nil
}

// formatServiceID is a method that validates the session service id
func (s *Session) formatServiceID(repo port.Repository, filled bool) error {
	if s.ServiceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyServiceID)
	}
	service := &Service{ID: s.ServiceID}
	if ok, err := service.Load(repo); err != nil {
		return err
	} else if !ok {
		return errors.New(pkg.ErrServiceNotFound)
	}
	return nil
}

// formatAt is a method that validates the session at
func (s *Session) formatAt(filled bool) error {
	if s.At.IsZero() {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyAt)
	}
	return nil
}


// formatKind is a method that validates the session kind
func (s *Session) formatKind(filled bool) error {
	if s.Kind == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyKind)
	}
	if !slices.Contains(KindSession, s.Kind) {
		return errors.New(pkg.ErrInvalidKind)
	}
	return nil
}

// formatStatus is a method that validates the session status
func (s *Session) formatStatus(filled bool) error {
	if s.Status == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyStatus)
	}
	if !slices.Contains(StatusSession, s.Status) {
		return errors.New(pkg.ErrInvalidStatus)
	}
	return nil
}

// validateDuplicity is a method that validates the duplicity of a client
func (s *Session) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Session{}, s.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, s.ID)
	}
	return nil
}