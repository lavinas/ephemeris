package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaCrud represents the dto for getting a agenda
type AgendaCrud struct {
	Base
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort       string `json:"sort" command:"name:sort;pos:3+"`
	Csv        string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID         string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date       string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	ClientID   string `json:"client" command:"name:client;pos:3+;trans:client_id,string" csv:"client"`
	ServiceID  string `json:"service" command:"name:service;pos:3+;trans:service_id,string" csv:"service"`
	ContractID string `json:"contract" command:"name:contract;pos:3+;trans:contract_id,string" csv:"contract"`
	Start      string `json:"start" command:"name:start;pos:3+;trans:start,time" csv:"start"`
	End        string `json:"end" command:"name:end;pos:3+;trans:end,time" csv:"end"`
	Price      string `json:"price" command:"name:price;pos:3+;trans:price,float" csv:"price"`
	Kind       string `json:"kind" command:"name:kind;pos:3+;trans:kind,string" csv:"kind"`
	Status     string `json:"status" command:"name:status;pos:3+;trans:status,string" csv:"status"`
	Bond       string `json:"bond" command:"name:bond;pos:3+;trans:bond,string" csv:"bond"`
	Billing    string `json:"billing" command:"name:billing;pos:3+;trans:billing_month,time" csv:"billing"`
}

// Validate is a method that validates the dto
func (a *AgendaCrud) Validate(repo port.Repository) error {
	if a.Csv != "" && (a.ID != "" || a.Date != "" || a.ClientID != "" || a.ContractID != "" || a.Start != "" || a.End != "" ||
		a.Kind != "" || a.Status != "" || a.Bond != "" || a.Billing != "" || a.Price != "" || a.ServiceID != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (a *AgendaCrud) GetCommand() string {
	return a.Action
}

// GetDomain is a method that returns a string representation of the agenda
func (a *AgendaCrud) GetDomain() []port.Domain {
	if a.Csv != "" {
		domains := []port.Domain{}
		agendas := []*AgendaCrud{}
		a.ReadCSV(&agendas, a.Csv)
		for _, ag := range agendas {
			ag.Action = a.Action
			ag.Object = a.Object
			domains = append(domains, ag.getDomain(ag))
		}
		return domains
	}
	return []port.Domain{
		a.getDomain(a),
	}
}

// getDomain is a method that returns the domain of one object
func (a *AgendaCrud) getDomain(one *AgendaCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Kind == "" {
		one.Kind = pkg.DefaultAgendaKind
	}
	if one.Action == "add" && one.Status == "" {
		one.Status = pkg.DefaultAgendaStatus
	}
	return domain.NewAgenda(one.ID, one.Date, one.ClientID, one.ServiceID, one.ContractID, 
		one.Start, one.End, one.Price, one.Kind, one.Status, one.Bond, one.Billing)
}

// GetOut is a method that returns the output dto
func (a *AgendaCrud) GetOut() port.DTOOut {
	return a
}

// GetDTO is a method that returns the dto
func (a *AgendaCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		agenda := slice.(*[]domain.Agenda)
		for _, ag := range *agenda {
			bond := ""
			if ag.Bond != nil {
				bond = a.Bond
			}
			billing := ""
			if ag.BillingMonth != nil {
				billing = ag.BillingMonth.Format(pkg.DateFormat)
			}
			contractID := ""
			if ag.ContractID != nil {
				contractID = *ag.ContractID
			}
			price := ""
			if ag.Price != nil {
				price = fmt.Sprintf("%.2f", *ag.Price)
			}
			ret = append(ret, &AgendaCrud{
				ID:         ag.ID,
				Date:       ag.Date.Format(pkg.DateFormat),
				ClientID:   ag.ClientID,
				ServiceID:  ag.ServiceID,
				ContractID: contractID,
				Start:      ag.Start.Format(pkg.DateTimeFormat),
				End:        ag.End.Format(pkg.DateTimeFormat),
				Price:      price,
				Kind:       ag.Kind,
				Status:     ag.Status,
				Bond:       bond,
				Billing:    billing,
			})
		}
	}
	pkg.NewCommands().Sort(ret, a.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (a *AgendaCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return a.getInstructions(a, domain)
}
