package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Service represents the service entity
type Service struct {
	ID      string    `gorm:"type:varchar(50); primaryKey"`
	Date    time.Time `gorm:"type:datetime; not null; index"`
	Name    string    `gorm:"type:varchar(100); not null; index"`
	Minutes *int64    `gorm:"type:int;  null; index"`
}

// NewService is a function that creates a new service
func NewService(id, date, name, minutes string) *Service {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	var min *int64 = nil
	if m, _ := strconv.ParseInt(minutes, 10, 64); m > 0 {
		min = &m
	}
	return &Service{
		ID:      id,
		Date:    fdate,
		Name:    name,
		Minutes: min,
	}
}

// Format is a method that formats the service
func (s *Service) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
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
	tx := repo.Begin()
	defer repo.Rollback(tx)
	if err := s.validateDuplicity(repo, tx, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return errors.New(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if a service exists
func (s *Service) Load(repo port.Repository) (bool, error) {
	tx := repo.Begin()
	defer repo.Rollback(tx)
	return repo.Get(tx, s, s.ID, false)
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
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	if len(s.ID) > 50 {
		return errors.New(pkg.ErrLongID50)
	}
	if len(strings.Split(s.ID, " ")) > 1 {
		return errors.New(pkg.ErrInvalidID)
	}
	return nil
}

// formatDate is a method that formats the service date
func (s *Service) formatDate(filled bool) error {
	if filled {
		return nil
	}
	if s.Date.IsZero() {
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
	}
	return nil
}

// formatName is a method that formats the service name
func (s *Service) formatName(filled bool) error {
	s.Name = s.formatString(s.Name)
	if filled {
		return nil
	}
	if s.Name == "" {
		return errors.New(pkg.ErrEmptyName)
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

// validateDuplicity is a method that validates the duplicity of a client
func (c *Service) validateDuplicity(repo port.Repository, tx interface{}, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(tx, &Service{}, c.ID, false)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}
