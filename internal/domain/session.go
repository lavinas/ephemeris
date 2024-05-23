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
	StatusSession = []string{pkg.SessionStatusDone, pkg.SessionStatusSaved, pkg.SessionStatusMissed}
	StatusProcess = []string{pkg.ProcessStatusWait, pkg.ProcessStatusSuccess, pkg.ProcessStatusError}
)

// Session represents the session entity
type Session struct {
	ID        string    `gorm:"type:varchar(150); primaryKey"`
	Date      time.Time `gorm:"type:datetime; not null"`
	ClientID  string    `gorm:"type:varchar(50); not null; index"`
	ServiceID string    `gorm:"type:varchar(50); not null; index"`
	At        time.Time `gorm:"type:datetime; not null"`
	Status    string    `gorm:"type:varchar(50); not null; index"`
	Discount  *float64  `gorm:"type:decimal(5,4); not null"`
	Process   string    `gorm:"type:varchar(50); not null; index"`
	Message   string    `gorm:"type:varchar(255); not null"`
}

// NewSession creates a new session domain entity
func NewSession(id, date, clientID, serviceID, at, status string, discount, process, message string) *Session {
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
	session.Status = status
	if disc, err := strconv.ParseFloat(discount, 64); err == nil {
		session.Discount = &disc
	}
	session.Process = process
	session.Message = message
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
	if err := s.formatServiceID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatAt(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatDiscount(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatProcess(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatMessage(filled); err != nil {
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
	if len(s.ID) > 150 {
		return errors.New(pkg.ErrLongID150)
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

// formatStatus is a method that validates the session status
func (s *Session) formatStatus(filled bool) error {
	if s.Status == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyStatus)
	}
	if !slices.Contains(StatusSession, s.Status) {
		status := strings.Join(StatusSession, ", ")
		return fmt.Errorf(pkg.ErrInvalidStatus, status)
	}
	return nil
}

// formatDiscount is a method that validates the session discount
func (s *Session) formatDiscount(filled bool) error {
	if s.Discount == nil {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyDiscount)
	}
	if *s.Discount < 0 || *s.Discount > 1 {
		return errors.New(pkg.ErrInvalidDiscount)
	}
	return nil
}

// formatProcess is a method that validates the session process
func (s *Session) formatProcess(filled bool) error {
	if s.Process == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyProcess)
	}
	if !slices.Contains(StatusProcess, s.Process) {
		status := strings.Join(StatusProcess, ", ")
		return fmt.Errorf(pkg.ErrInvalidProcess, status[:len(status)-2])
	}
	return nil
}

// formatMessage is a method that validates the session message
func (s *Session) formatMessage(filled bool) error {
	if s.Process == pkg.ProcessStatusError && s.Message == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyMessage)
	}
	if len(s.Message) > 255 {
		return errors.New(pkg.ErrLongMessage255)
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
