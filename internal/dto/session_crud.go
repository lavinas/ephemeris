package dto

import (
	"errors"

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


// isEmpty is a method that checks if the dto is empty
func (s *SessionCrud) isEmpty() bool {
	return s.ID == "" && s.Date == "" && s.ClientID == "" && s.ContractID == "" && s.Minutes == "" && s.Kind == "" && s.Status == ""
}