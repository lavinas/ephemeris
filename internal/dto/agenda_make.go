package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// AgendaMake represents the dto for making a agenda
type AgendaMake struct {
	Object     string `json:"-" command:"name:agenda;key;pos:2-"`
	Action     string `json:"-" command:"name:make,program,prog;key;pos:2-"`
	ClientID   string `json:"client_id" command:"name:client;pos:3+"`
	Month      string `json:"month" command:"name:month;pos:3+"`
}

// AgendaMakeOut represents the dto for making a agenda on output
type AgendaMakeOut struct {
	ID         string `json:"id" command:"name:id"`
	ClientID   string `json:"client_id" command:"name:client"`
	ServiceID  string `json:"service_id" command:"name:service"`
	ContractID string `json:"contract_id" command:"name:contract"`
	Start      string `json:"start" command:"name:start"`
	End        string `json:"end" command:"name:end"`
	Event      string `json:"event" command:"name:event"`
	Price      string `json:"price" command:"name:price"`
	Kind       string `json:"kind" command:"name:kind"`
	Status     string `json:"status" command:"name:status"`
}

// Validate is a method that validates the dto
func (a *AgendaMake) Validate(repo port.Repository) error {
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
	return []port.Domain{
		&domain.Contract{
			ClientID:  a.ClientID,
		},
	}
}

// GetOut is a method that returns the dto out
func (a *AgendaMake) GetOut() port.DTOOut {
	return &AgendaMakeOut{}
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (a *AgendaMake) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	month, err := time.Parse(pkg.MonthFormat, a.Month)
	if err != nil {
		return nil, nil, err
	}
	firstday := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	lastday := firstday.AddDate(0, 1, 0).Add(time.Nanosecond * -1)
	p1 := fmt.Sprintf("start <= '%s'", lastday.Format("2006-01-02 15:04:05"))
	p2 := fmt.Sprintf("end is null or end >= '%s'", firstday.Format("2006-01-02 15:04:05"))
	return nil, []interface{}{p1, p2}, nil
}

// GetDTO is a method that returns the dto out
func (a *AgendaMakeOut) GetDTO(domainIn interface{}) []port.DTOOut {
	agenda := domainIn.(*domain.Agenda)
	contract := ""
	if agenda.ContractID != nil {
		contract = *agenda.ContractID
	}
	price := ""
	if agenda.Price != nil {
		price = fmt.Sprintf("%.2f", *agenda.Price)
	}
	return []port.DTOOut{
		&AgendaMakeOut{
			ID:         agenda.ID,
			ClientID:   agenda.ClientID,
			ServiceID:  agenda.ServiceID,
			ContractID: contract,
			Start:      agenda.Start.Format(pkg.DateTimeFormat),
			End:        agenda.End.Format(pkg.DateTimeFormat),
			Price:      price,
			Kind:       agenda.Kind,
			Status:     agenda.Status,
		},
	}
}
