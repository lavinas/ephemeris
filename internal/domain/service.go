package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
)

// Service represents the service entity
type Service struct {
	ID   string    `gorm:"type:varchar(25); primaryKey"`
	Date time.Time `gorm:"type:datetime; not null"`
	Name string    `gorm:"type:varchar(100), not null"`
}

// NewService is a function that creates a new service
func NewService(id string, date string, name string) *Service {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(port.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(port.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	return &Service{
		ID:   id,
		Date: fdate,
		Name: name,
	}
}

// Format is a method that formats the service
func (s *Service) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	msg := ""
	if err := s.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := s.formatName(filled); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg)
	}
	return nil
}

// GetID is a method that returns the id of the client
func (s *Service) GetID() string {
	return s.ID
}

// Get is a method that returns the client
func (s *Service) Get() port.Domain {
	return s
}

// GetEmpty is a method that returns an empty client with just id
func (c *Service) GetEmpty() port.Domain {
	return &Service{}
}

// TableName returns the table name for database
func (b *Service) TableName() string {
	return "service"
}

// formatID is a method that formats the service id
func (s *Service) formatID(filled bool) error {
	s.ID = s.formatString(s.ID)
	if s.ID == "" {
		if !filled {
			return nil
		}
		return errors.New("service id is required")
	}
	if len(s.ID) > 25 {
		return errors.New("service id must have at most 25 characters")
	}
	if len(strings.Split(s.ID, " ")) > 1 {
		return errors.New(port.ErrInvalidID)
	}
	return nil
}

// formatDate is a method that formats the service date
func (s *Service) formatDate(filled bool) error {
	if !filled {
		return nil
	}
	if s.Date.IsZero() {
		return errors.New("service date is required")
	}
	return nil
}

// formatName is a method that formats the service name
func (s *Service) formatName(filled bool) error {
	s.Name = s.formatString(s.Name)
	if !filled {
		return nil
	}
	if s.Name == "" {
		return errors.New("service name is required")
	}
	return nil
}

// formatString is a method that formats a string
func (s *Service) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}
