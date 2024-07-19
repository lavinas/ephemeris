package dto

import (
	"errors"
	"strconv"
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type InvoiceItemCrud struct {
	Base
	Object      string `json:"-" command:"name:item;key;pos:2-"`
	Action      string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort        string `json:"sort" command:"name:sort;pos:3+"`
	Csv         string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID          string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	InvoiceID   string `json:"invoice" command:"name:invoice;pos:3+;trans:invoice_id,string" csv:"invoice"`
	AgendaID    string `json:"agenda" command:"name:agenda;pos:3+;trans:agenda_id,string" csv:"agenda"`
	Value       string `json:"value" command:"name:value;pos:3+;trans:value,numeric" csv:"value"`
	Description string `json:"description" command:"name:description;pos:3+;trans:description,string" csv:"description"`
}

// Validate is a method that validates the dto
func (i *InvoiceItemCrud) Validate() error {
	if i.Csv != "" && (i.ID != "" || i.InvoiceID != "" || i.AgendaID != "" || i.Value != "" || i.Description != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (i *InvoiceItemCrud) GetCommand() string {
	return i.Action
}

// GetDomain is a method that returns a string representation of the invoice item
func (i *InvoiceItemCrud) GetDomain() []port.Domain {
	if i.Csv != "" {
		domains := []port.Domain{}
		items := []*InvoiceItemCrud{}
		i.ReadCSV(&items, i.Csv)
		for _, item := range items {
			item.Action = i.Action
			item.Object = i.Object
			domains = append(domains, i.getDomain(item))
		}
		return domains
	}
	return []port.Domain{i.getDomain(i)}
}

// GetOut is a method that returns the output dto
func (i *InvoiceItemCrud) GetOut() port.DTOOut {
	return i
}

// GetDTO is a method that returns the dto for given domain
func (i *InvoiceItemCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		items := slice.(*[]domain.InvoiceItem)
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
	}
	pkg.NewCommands().Sort(ret, i.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (i *InvoiceItemCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return i.getInstructions(i, domain)
}

// getDomain is a method that returns a string representation of the invoice item
func (i *InvoiceItemCrud) getDomain(one *InvoiceItemCrud) port.Domain {
	i.trim()
	return domain.NewInvoiceItem(one.ID, one.InvoiceID, one.AgendaID, one.Value, one.Description)
}

// trim is a method that trims the dto
func (i *InvoiceItemCrud) trim() {
	i.ID = strings.TrimSpace(i.ID)
	i.InvoiceID = strings.TrimSpace(i.InvoiceID)
	i.AgendaID = strings.TrimSpace(i.AgendaID)
	i.Value = strings.TrimSpace(i.Value)
	i.Description = strings.TrimSpace(i.Description)
}
