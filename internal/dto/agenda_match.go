package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMatch represents the dto for matching a agenda with sessions
type AgendaMatch struct {
	Object   string `json:"-" command:"name:agenda;key;pos:2-"`
	Action   string `json:"-" command:"name:match;key;pos:2-"`
	ClientID string `json:"client_id" command:"name:client;pos:3+"`
	Month    string `json:"month" command:"name:month;pos:3+"`
}

// Validate is a method that validates the dto
func (a *AgendaMatch) Validate(repo port.Repository) error {
	if _, err := a.parseMonth(); err != nil {
		return err
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (a *AgendaMatch) GetCommand() string {
	return a.Action
}

// GetDomain is a method that returns the domain of the dto
func (a *AgendaMatch) GetDomain() []port.Domain {
	return []port.Domain{
		&domain.Agenda{
			ClientID: a.ClientID,
			Kind:     pkg.AgendaKindRegular,
			Status:   pkg.AgendaStatusOpen,
		},
	}
}

// GetOut is a method that returns the output dto
func (a *AgendaMatch) GetOut() port.DTOOut {
	return a
}

// GetInstructions is a method that returns the instructions of the dto
func (a *AgendaMatch) GetInstructions() string {
	month, _ := a.parseMonth()
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	return fmt.Sprintf("At >= '%s' and At <= '%s'", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))
}

// GetDTO is a method that returns the dto
func (a *AgendaMatch) GetDTO(domainIn interface{}) []port.DTOOut {
	return []port.DTOOut{a}
}

// parseMonth is a method that parses the month
func (a *AgendaMatch) parseMonth() (time.Time, error) {
	if a.Month == "" {
		return time.Time{}, errors.New(pkg.ErrMonthEmpty)
	}
	return time.Parse(pkg.MonthFormat, a.Month)
}
