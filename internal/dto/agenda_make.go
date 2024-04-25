package dto

import (
	"errors"

	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMake represents the dto for making a agenda
type AgendaMake struct {
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:create;key;pos:2-"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+"`
	ContractID string `json:"contract_id" command:"name:contract;pos:3+"`
	Month      string `json:"month" command:"name:month;pos:3+"`
}


// AgendaMakeOut represents the dto for making a agenda on output
type AgendaMakeOut struct {
	ID         string `json:"id"`
	ClientID   string `json:"client_id"`
	ContractID string `json:"contract_id"`
	Start	   string `json:"start"`
	End 	   string `json:"end"`
}

// Validate is a method that validates the dto
func (a *AgendaMake) Validate() error {
	if a.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (a *AgendaMake) GetCommand() string {
	return a.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (a *AgendaMake) GetDomain() []interface{} {
	return []interface{}{a}
}

// GetOut is a method that returns the output dto
func (a *AgendaMake) GetOut() interface{} {
	return &AgendaMakeOut{}
}

// GetDTO is a method that returns the dto
func (a *AgendaMake) GetDTO(domainIn interface{}) []interface{} {
	ret := []interface{}{}
	slices := domainIn.([]interface{})
	return append(ret, slices...)
}

// isEmpty is a method that returns true if the dto is empty
func (a *AgendaMake) isEmpty() bool {
	return a.ClientID == "" || a.ContractID == "" || a.Month == ""
}

	
