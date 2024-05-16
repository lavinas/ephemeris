package domain

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
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
	ContractID string    `gorm:"type:varchar(50); index"`
	At         time.Time `gorm:"type:datetime; not null"`
	Minutes    int       `gorm:"type:int; not null"`
	Kind       string    `gorm:"type:varchar(50); not null; index"`
	Status     string    `gorm:"type:varchar(50); not null; index"`
}

// NewSession creates a new session domain entity
func NewSession(id, date, clientID, contractID, at, minutes, kind, status string) *Session {
	session := &Session{}
	session.ID = id
	session.ClientID = clientID
	session.ContractID = contractID
	local, _ := time.LoadLocation(pkg.Location)
	session.Date, _ = time.ParseInLocation(pkg.DateFormat, date, local)
	session.At, _ = time.ParseInLocation("2006-01-02T15:04:05", at, local)
	session.Minutes, _ = strconv.Atoi(minutes)
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
	if err := s.formatClientContractID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatAt(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatMinutes(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatKind(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
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
func (s *Session) formatClientContractID(repo port.Repository, filled bool) error {
	if s.ClientID == "" && s.ContractID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyClientOrContractID)
	}
	if s.ContractID != "" {
		contract := &Contract{ID: s.ContractID}
		if ok, err := contract.Load(repo); err != nil {
			return err
		} else if !ok {
			return errors.New(pkg.ErrContractNotFound)
		}
		if s.ClientID != "" && contract.ClientID != s.ClientID {
			return errors.New(pkg.ErrContractClientMismatch)
		}
	}
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

// formatMinutes is a method that validates the session minutes
func (s *Session) formatMinutes(filled bool) error {
	if s.Minutes == 0 {
		if filled {
			return nil
		}
		if s.Kind != pkg.SessionKindRegular {
			return errors.New(pkg.ErrEmptyMinutes)
		}
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
