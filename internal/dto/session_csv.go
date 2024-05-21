package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionCSV represents the dto for getting a session
type SessionCSV struct {
	Object     string `json:"-" command:"name:session;key;pos:2-"`
	Action     string `json:"-" command:"name:csv;key;pos:2-"`
	File       string `json:"file" command:"name:file;pos:3+;"`
}

type SessionCSVOut struct {
	ClientID   string `csv:"client"`
	ServiceID  string `csv:"service"`
	At         string `csv:"at"`
	Kind       string `csv:"kind"`
	Status     string `csv:"status"`
	Result     string `csv:"result"`
	Message    string `csv:"message"`
}


// SessionCSVLine represents the dto for getting a session
type SessionCSVLine struct {
	ServiceCrud
}

// Validate is a method that validates the dto
func (s *SessionCSV) Validate(repo port.Repository) error {
	if s.File == "" {
		return errors.New(pkg.ErrFileNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *SessionCSV) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (s *SessionCSV) GetDomain() []port.Domain {
	return []port.Domain{}
}

// GetOut is a method that returns the output dto
func (s *SessionCSV) GetOut() port.DTOOut {
	return &SessionCSVLine{}
}

// GetDTO is a method that returns the dto
func (s *SessionCSVLine) GetDTO(domainIn interface{}) []port.DTOOut {
	return []port.DTOOut{}
}

// GetInstructions is a method that returns the instructions of the dto
func (s *SessionCSV) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return nil, nil, nil
}


