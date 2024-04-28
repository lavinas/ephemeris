package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMake represents the dto for making a agenda
type AgendaMake struct {
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:make;key;pos:2-"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+"`
	ContractID string `json:"contract_id" command:"name:contract;pos:3+"`
	Month      string `json:"month" command:"name:month;pos:3+"`
}

// AgendaMakeOut represents the dto for making a agenda on output
type AgendaMakeOut struct {
	ID         string `json:"id" command:"name:id"`
	ClientID   string `json:"client_id" command:"name:client"`
	ContractID string `json:"contract_id" command:"name:contract"`
	Start      string `json:"start" command:"name:start"`
	End        string `json:"end" command:"name:end"`
}

// Validate is a method that validates the dto
func (a *AgendaMake) Validate(repo port.Repository) error {
	if a.ContractID == "" && a.ClientID == "" {
		return errors.New(pkg.ErrClientContractEmpty)
	}
	if a.Month == "" {
		return errors.New(pkg.ErrMonthEmpty)
	}
	if _, err := time.Parse(pkg.MonthFormat, a.Month); err != nil {
		return fmt.Errorf(pkg.ErrMonthInvalid, pkg.MonthFormat)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (a *AgendaMake) GetCommand() string {
	return a.Action
}

// GetDomain is a method that returns the domain of the dto
func (a *AgendaMake) GetDomain() []port.Domain {
	return []port.Domain{}
}

// GetOut is a method that returns the dto out
func (a *AgendaMake) GetOut() port.DTOOut {
	return &AgendaMakeOut{}
}

// GetDTO is a method that returns the dto out
func (a *AgendaMakeOut) GetDTO(domainIn interface{}) []port.DTOOut {
	return []port.DTOOut{}
}
