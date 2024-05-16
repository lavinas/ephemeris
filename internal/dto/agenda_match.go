package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
	"github.com/lavinas/ephemeris/internal/domain"
)

// AgendaMatch represents the dto for matching a agenda with sessions
type AgendaMatch struct {
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:match;key;pos:2-"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+"`
	ContractID string `json:"contract_id" command:"name:contract;pos:3+"`
	Month      string `json:"month" command:"name:month;pos:3+"`
}

// Validate is a method that validates the dto
func (a *AgendaMatch) Validate(repo port.Repository) error {
	if a.Month == "" {
		return errors.New(pkg.ErrMonthEmpty)
	}
	if _, err := time.Parse(pkg.MonthFormat, a.Month); err != nil {
		return fmt.Errorf(pkg.ErrMonthInvalid, pkg.MonthFormat)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (a *AgendaMatch) GetCommand() string {
	return a.Action
}

// GetDomain is a method that returns the domain of the dto
func (a *AgendaMatch) GetDomain() []port.Domain {
	time.Local, _ = time.LoadLocation(pkg.Location)
	d := time.Now()
	date := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.Local)
	return []port.Domain{
		&domain.Agenda{
			Date:       date,
			ContractID: a.ContractID,
			Kind:       pkg.AgendaKindRegular,
			Status:     pkg.AgendaStatusOpen,
		},
	}
}

// GetOut is a method that returns the output dto
func (a *AgendaMatch) GetOut() port.DTOOut {
	return a
}

// GetInstructions is a method that returns the instructions of the dto
func (a *AgendaMatch) GetInstructions() string {
	return fmt.Sprintf("Matching agenda for client %s and contract %s on month %s", a.ClientID, a.ContractID, a.Month)
}

// GetDTO is a method that returns the dto
func (a *AgendaMatch) GetDTO(domainIn interface{}) []port.DTOOut {
	return []port.DTOOut{a}
}



