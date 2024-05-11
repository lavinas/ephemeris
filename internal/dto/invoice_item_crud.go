package dto

import (
	"errors"
	"strconv"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type InvoiceItemCrud struct {
	Object      string `json:"-" command:"name:item;key;pos:2-"`
	Action      string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort        string `json:"sort" command:"name:sort;pos:3+"`
	ID          string `json:"id" command:"name:id;pos:3+;trans:id,string"`
	InvoiceID   string `json:"invoice_id" command:"name:invoice;pos:3+;trans:invoice_id,string"`
	AgendaID    string `json:"agenda_id" command:"name:agenda;pos:3+;trans:agenda_id,string"`
	Value       string `json:"value" command:"name:value;pos:3+;trans:value,numeric"`
	Description string `json:"description" command:"name:description;pos:3+;trans:description,string"`
}

// Validate is a method that validates the dto
func (i *InvoiceItemCrud) Validate(repo port.Repository) error {
	if i.Action != "get" && i.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (i *InvoiceItemCrud) GetCommand() string {
	return i.Action
}

// GetDomain is a method that returns a string representation of the invoice item
func (i *InvoiceItemCrud) GetDomain() []port.Domain {
	return []port.Domain{
		domain.NewInvoiceItem(i.ID, i.InvoiceID, i.AgendaID, i.Value, i.Description),
	}
}

// GetOut is a method that returns the output dto
func (i *InvoiceItemCrud) GetOut() port.DTOOut {
	return i
}

func (i *InvoiceItemCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	items := slices[0].(*[]domain.InvoiceItem)
	for _, item := range *items {
		agendaID := ""
		if item.AgendaID != nil {
			agendaID = *item.AgendaID
		}
		ret = append(ret, &InvoiceItemCrud{
			ID:          item.ID,
			InvoiceID:   item.InvoiceID,
			AgendaID:    agendaID,
			Value:       strconv.FormatFloat(item.Value, 'f', 2, 64),
			Description: item.Description,
		})
	}
	pkg.NewCommands().Sort(ret, i.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (i *InvoiceItemCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	cmd, err := pkg.NewCommands().Transpose(i)
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) > 0 {
		domain := i.GetDomain()[0]
		return domain, cmd, nil
	}
	return domain, cmd, nil
}

// isEmpty is a method that checks if the dto is empty
func (i *InvoiceItemCrud) isEmpty() bool {
	if i.ID == "" && i.InvoiceID == "" && i.AgendaID == "" && i.Value == "" && i.Description == "" {
		return true
	}
	return false
}
