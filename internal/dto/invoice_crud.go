package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type InvoiceCrud struct {
	Object        string `json:"-" command:"name:invoice;key;pos:2-"`
	Action        string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort          string `json:"sort" command:"name:sort;pos:3+"`
	ID            string `json:"id" command:"name:id;pos:3+;trans:id,string"`
	Date          string `json:"date" command:"name:date;pos:3+;trans:date,time"`
	ClientID      string `json:"client_id" command:"name:client;pos:3+;trans:client_id,string"`
	Value         string `json:"value" command:"name:value;pos:3+;trans:value,numeric"`
	Status        string `json:"status" command:"name:status;pos:3+;trans:status,string"`
	SendStatus    string `json:"send_status" command:"name:send_status;pos:3+;trans:send_status,string"`
	PaymentStatus string `json:"payment_status" command:"name:payment_status;pos:3+;trans:payment_status,string"`
}

// Validate is a method that validates the dto
func (i *InvoiceCrud) Validate(repo port.Repository) error {
	if i.Action != "get" && i.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (i *InvoiceCrud) GetCommand() string {
	return i.Action
}

// GetDomain is a method that returns a string representation of the invoice
func (i *InvoiceCrud) GetDomain() []port.Domain {
	if i.Action == "add" && i.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		i.Date = time.Now().Format(pkg.DateFormat)
	}
	if i.Action == "add" && i.Status == "" {
		i.Status = pkg.DefaultInvoiceStatus
	}
	if i.Action == "add" && i.SendStatus == "" {
		i.SendStatus = pkg.DefaultInvoiceSendStatus
	}
	if i.Action == "add" && i.PaymentStatus == "" {
		i.PaymentStatus = pkg.DefaultInvoicePaymentStatus
	}
	return []port.Domain{
		domain.NewInvoice(i.ID, i.ClientID, i.Date, i.Value, i.Status, i.SendStatus, i.PaymentStatus),
	}
}

// GetOut is a method that returns the output dto
func (i *InvoiceCrud) GetOut() port.DTOOut {
	return i
}

// GetDTO is a method that returns the output dto
func (i *InvoiceCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	invoices := slices[0].(*[]domain.Invoice)
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
	pkg.NewCommands().Sort(ret, i.Sort)
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (i *InvoiceCrud) isEmpty() bool {
	return i.ID == "" && i.Date == "" && i.ClientID == "" && i.Value == "" && i.Status == "" && i.SendStatus == "" && i.PaymentStatus == ""
}
