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
	ID          string `json:"id" command:"name:id;pos:3+"`
	InvoiceID   string `json:"invoice_id" command:"name:invoice;pos:3+"`
	AgendaID    string `json:"agenda_id" command:"name:agenda;pos:3+"`
	Value       string `json:"value" command:"name:value;pos:3+"`
	Description string `json:"description" command:"name:description;pos:3+"`
}

// Validate is a method that validates the dto
func (i *InvoiceItemCrud) Validate(repo port.Repository) error {
	if i.isEmpty() {
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
	return &InvoiceItemCrud{}
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
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (i *InvoiceItemCrud) isEmpty() bool {
	if i.ID == "" && i.InvoiceID == "" && i.AgendaID == "" && i.Value == "" && i.Description == "" {
		return true
	}
	return false
}
