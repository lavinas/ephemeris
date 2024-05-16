package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// SessionCrud represents the dto for getting a session
type SessionCrud struct {
	Object     string `json:"-" command:"name:session;key;pos:2-"`
	Action     string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort       string `json:"sort" command:"name:sort;pos:3+"`
	ID         string `json:"id" command:"name:id;pos:3+;trans:id,string"`
	Date       string `json:"date" command:"name:date;pos:3+;trans:date,time"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+;trans:client_id,string"`
	ContractID string `json:"contract_id" command:"name:contract;pos:3+;trans:contract_id,string"`
	At         string `json:"at" command:"name:at;pos:3+;trans:at,time"`
	Minutes    string `json:"minutes" command:"name:minutes;pos:3+;trans:minutes,numeric"`
	Kind       string `json:"kind" command:"name:kind;pos:3+;trans:kind,string"`
	Status     string `json:"status" command:"name:status;pos:3+;trans:status,string"`
}

// Validate is a method that validates the dto
func (s *SessionCrud) Validate(repo port.Repository) error {
	if s.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (s *SessionCrud) GetCommand() string {
	return s.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (s *SessionCrud) GetDomain() []port.Domain {
	if s.Action == "add" && s.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		s.Date = time.Now().Format(pkg.DateFormat)
	}
	if s.Action == "add" && s.Kind == "" {
		s.Kind = pkg.DefaultSessionKind
	}
	if s.Action == "add" && s.Status == "" {
		s.Status = pkg.DefaultSessionStatus
	}
	return []port.Domain{
		domain.NewSession(s.ID, s.Date, s.ClientID, s.ContractID, s.At, s.Minutes, s.Kind, s.Status),
	}
}

// GetOut is a method that returns the output dto
func (s *SessionCrud) GetOut() port.DTOOut {
	return s
}

// GetDTO is a method that returns the dto
func (s *SessionCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	sessions := slices[0].(*[]domain.Session)
	for _, se := range *sessions {
		ret = append(ret, &SessionCrud{
			ID:         se.ID,
			Date:       se.Date.Format(pkg.DateFormat),
			ClientID:   se.ClientID,
			ContractID: se.ContractID,
			At:         se.At.Format(pkg.DateTimeFormat),
			Minutes:    strconv.FormatInt(se.Minutes, 10),
			Kind:       se.Kind,
			Status:     se.Status,
		})
	}
	pkg.NewCommands().Sort(ret, s.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (s *SessionCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	cmd, err := pkg.NewCommands().Transpose(s)
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) > 0 {
		domain := s.GetDomain()[0]
		return domain, cmd, nil
	}
	return domain, cmd, nil
}

// isEmpty is a method that checks if the dto is empty
func (s *SessionCrud) isEmpty() bool {
	return s.ID == "" && s.Date == "" && s.ClientID == "" && s.ContractID == "" && s.Minutes == "" && s.Kind == "" && s.Status == ""
}