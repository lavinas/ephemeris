package dto

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type InvoiceCrud struct {
	Base
	Object        string `json:"-" command:"name:invoice;key;pos:2-"`
	Action        string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort          string `json:"sort" command:"name:sort;pos:3+"`
	Csv           string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID            string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date          string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	ClientID      string `json:"client" command:"name:client;pos:3+;trans:client_id,string" csv:"client"`
	Value         string `json:"value" command:"name:value;pos:3+;trans:value,numeric" csv:"value"`
	Status        string `json:"status" command:"name:status;pos:3+;trans:status,string" csv:"status"`
	SendStatus    string `json:"send_status" command:"name:send_status;pos:3+;trans:send_status,string" csv:"send_status"`
	PaymentStatus string `json:"payment_status" command:"name:payment_status;pos:3+;trans:payment_status,string" csv:"payment_status"`
}

// Validate is a method that validates the dto
func (i *InvoiceCrud) Validate() error {
	if i.Csv != "" && (i.ID != "" || i.Date != "" || i.ClientID != "" || i.Value != "" || i.Status != "" || i.SendStatus != "" || i.PaymentStatus != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (i *InvoiceCrud) GetCommand() string {
	return i.Action
}

// GetDomain is a method that returns a string representation of the invoice
func (i *InvoiceCrud) GetDomain() []port.Domain {
	if i.Csv != "" {
		domains := []port.Domain{}
		invoices := []*InvoiceCrud{}
		i.ReadCSV(&invoices, i.Csv)
		for _, invoice := range invoices {
			invoice.Action = i.Action
			invoice.Object = i.Object
			domains = append(domains, i.getDomain(invoice))
		}
		return domains
	}
	return []port.Domain{i.getDomain(i)}
}

// GetOut is a method that returns the output dto
func (i *InvoiceCrud) GetOut() port.DTOOut {
	return i
}

// GetDTO is a method that returns the output dto
func (i *InvoiceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		invoices := slice.(*[]domain.Invoice)
		for _, invoice := range *invoices {
			ret = append(ret, &InvoiceCrud{
				ID:            invoice.ID,
				Date:          invoice.Date.Format(pkg.DateFormat),
				ClientID:      invoice.ClientID,
				Value:         strconv.FormatFloat(invoice.Value, 'f', 2, 64),
				Status:        invoice.Status,
				SendStatus:    invoice.SendStatus,
				PaymentStatus: invoice.PaymentStatus,
			})
		}
	}
	pkg.NewCommands().Sort(ret, i.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (i *InvoiceCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return i.getInstructions(i, domain)
}

// getDomain is a method that returns a string representation of the invoice
func (i *InvoiceCrud) getDomain(one *InvoiceCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Status == "" {
		one.Status = pkg.DefaultInvoiceStatus
	}
	if one.Action == "add" && one.SendStatus == "" {
		one.SendStatus = pkg.DefaultInvoiceSendStatus
	}
	if one.Action == "add" && one.PaymentStatus == "" {
		one.PaymentStatus = pkg.DefaultInvoicePaymentStatus
	}
	one.trim()
	return domain.NewInvoice(one.ID, one.ClientID, one.Date, one.Value, one.Status, one.SendStatus, one.PaymentStatus)
}

// trim is a method that trims the dto
func (i *InvoiceCrud) trim() {
	i.ID = strings.TrimSpace(i.ID)
	i.Date = strings.TrimSpace(i.Date)
	i.ClientID = strings.TrimSpace(i.ClientID)
	i.Value = strings.TrimSpace(i.Value)
	i.Status = strings.TrimSpace(i.Status)
	i.SendStatus = strings.TrimSpace(i.SendStatus)
	i.PaymentStatus = strings.TrimSpace(i.PaymentStatus)
}
