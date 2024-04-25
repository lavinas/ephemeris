package dto

import (
	"errors"
	"fmt"
	"strconv"
	"time"

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
	if a.ContractID == "" && a.ClientID == "" {
		return errors.New(pkg.ErrClientContractEmpty)
	}
	if a.Month == "" {
		return errors.New(pkg.ErrMonthEmpty)
	}
	if _, err := time.Parse(a.Month, pkg.MonthFormat); err != nil {
		return fmt.Errorf(pkg.ErrMonthInvalid, pkg.MonthFormat)
	}
	if _, err := strconv.ParseInt(a.ClientID, 10, 64); err != nil {
		return errors.New(pkg.ErrInvalidClientID)
	}

	return nil
}

// GetClient returns clientid from dto
func (a *AgendaMake) GetClientID() int {
	return 0
}

// GetContractID return contractID from dto
func (a *AgendaMake) GetContractID() int {
	return 0
}

// GetMonth returns month of dto
func (a *AgendaMake) GetMonth() time.Time {
	return time.Now()
}
	
