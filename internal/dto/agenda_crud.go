package dto

import (
	"errors"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaCrud represents the dto for getting a agenda
type AgendaCrud struct {
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID         string `json:"id" command:"name:id;pos:3+"`
	Date       string `json:"date" command:"name:date;pos:3+"`
	ContractID string `json:"contract_id" command:"name:contract;pos:3+"`
	Start      string `json:"start" command:"name:start;pos:3+"`
	End        string `json:"end" command:"name:end;pos:3+"`
	Kind       string `json:"kind" command:"name:kind;pos:3+"`
	Status     string `json:"status" command:"name:status;pos:3+"`
	Bond       string `json:"bond" command:"name:bond;pos:3+"`
	Billing    string `json:"billing" command:"name:billing;pos:3+"`
}

// Validate is a method that validates the dto
func (a *AgendaCrud) Validate(repo port.Repository) error {
	if a.isEmpty() && a.Action != "get" {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (a *AgendaCrud) GetCommand() string {
	return a.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (a *AgendaCrud) GetDomain() []port.Domain {
	if a.Action == "add" && a.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		a.Date = time.Now().Format(pkg.DateFormat)
	}
	if a.Action == "add" && a.Kind == "" {
		a.Kind = pkg.DefaulltAgendaKind
	}
	if a.Action == "add" && a.Status == "" {
		a.Status = pkg.DefaultAgendaStatus
	}
	return []port.Domain{
		domain.NewAgenda(a.ID, a.Date, a.ContractID, a.Start, a.End, a.Kind, a.Status, a.Bond, a.Billing),
	}
}

// GetOut is a method that returns the output dto
func (c *AgendaCrud) GetOut() port.DTOOut {
	return &AgendaCrud{}
}

// GetDTO is a method that returns the dto
func (c *AgendaCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	agenda := slices[0].(*[]domain.Agenda)
	for _, a := range *agenda {
		bond := ""
		if a.Bond != nil {
			bond = c.Bond
		}
		billing := ""
		if a.BillingMonth != nil {
			billing = a.BillingMonth.Format(pkg.DateFormat)
		}
		ret = append(ret, &AgendaCrud{
			ID:         a.ID,
			Date:       a.Date.Format(pkg.DateFormat),
			ContractID: a.ContractID,
			Start:      a.Start.Format(pkg.DateTimeFormat),
			End:        a.End.Format(pkg.DateTimeFormat),
			Kind:       a.Kind,
			Status:     a.Status,
			Bond:       bond,
			Billing:    billing,
		})
	}
	return ret
}

// isEmpty is a method that returns if the dto is empty
func (a *AgendaCrud) isEmpty() bool {
	return a.ID == "" && a.Date == "" && a.ContractID == "" && a.Start == "" && a.End == "" && a.Kind == "" && a.Status == "" && a.Bond == "" && a.Billing == ""
}
