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
	Sort       string `json:"sort" command:"name:sort;pos:3+"`
	ID         string `json:"id" command:"name:id;pos:3+;trans:id,string"`
	Date       string `json:"date" command:"name:date;pos:3+;trans:date,time"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+;trans:client_id,string"`
	ContractID string `json:"contract_id" command:"name:contract;pos:3+;trans:contract_id,string"`
	Start      string `json:"start" command:"name:start;pos:3+;trans:start,time"`
	End        string `json:"end" command:"name:end;pos:3+;trans:end,time"`
	Kind       string `json:"kind" command:"name:kind;pos:3+;trans:kind,string"`
	Status     string `json:"status" command:"name:status;pos:3+;trans:status,string"`
	Bond       string `json:"bond" command:"name:bond;pos:3+;trans:bond,string"`
	Billing    string `json:"billing" command:"name:billing;pos:3+;trans:billing_month,time"`
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
		domain.NewAgenda(a.ID, a.Date, a.ClientID, a.ContractID, a.Start, a.End, a.Kind, a.Status, a.Bond, a.Billing),
	}
}

// GetOut is a method that returns the output dto
func (a *AgendaCrud) GetOut() port.DTOOut {
	return a
}

// GetDTO is a method that returns the dto
func (a *AgendaCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	agenda := slices[0].(*[]domain.Agenda)
	for _, ag := range *agenda {
		bond := ""
		if ag.Bond != nil {
			bond = a.Bond
		}
		billing := ""
		if ag.BillingMonth != nil {
			billing = ag.BillingMonth.Format(pkg.DateFormat)
		}
		ret = append(ret, &AgendaCrud{
			ID:         ag.ID,
			Date:       ag.Date.Format(pkg.DateFormat),
			ClientID:   ag.ClientID,
			ContractID: ag.ContractID,
			Start:      ag.Start.Format(pkg.DateTimeFormat),
			End:        ag.End.Format(pkg.DateTimeFormat),
			Kind:       ag.Kind,
			Status:     ag.Status,
			Bond:       bond,
			Billing:    billing,
		})
	}
	pkg.NewCommands().Sort(ret, a.Sort)
	return ret
}

// isEmpty is a method that returns if the dto is empty
func (a *AgendaCrud) isEmpty() bool {
	return a.ID == "" && a.Date == "" && a.ContractID == "" && a.Start == "" && a.End == "" && a.Kind == "" && a.Status == "" && a.Bond == "" && a.Billing == ""
}
