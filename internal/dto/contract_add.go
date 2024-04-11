package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ContractAddIn represents the input dto for adding a contract usecase
type ContractAddIn struct {
	Object       string `json:"-" command:"name:contract;key"`
	Action       string `json:"-" command:"name:add;key"`
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	ClientID     string `json:"client" command:"name:client"`
	ServiceID    string `json:"service" command:"name:service"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence"`
	PriceID      string `json:"price" command:"name:price"`
	BillingType  string `json:"type" command:"name:billing"`
	DueDay       string `json:"due" command:"name:due"`
	Start        string `json:"start" command:"name:start"`
	Bond         string `json:"bond" command:"name:bond"`
}

// ContractAddOut represents the output dto for adding a contract usecase
type ContractAddOut struct {
	ID           string `json:"id" command:"name:id"`
	Date         string `json:"date" command:"name:date"`
	ClientID     string `json:"client" command:"name:client"`
	ServiceID    string `json:"service" command:"name:service"`
	RecurrenceID string `json:"recurrence" command:"name:recurrence"`
	PriceID      string `json:"price" command:"name:price"`
	BillingType  string `json:"type" command:"name:billing"`
	DueDay       string `json:"due" command:"name:due"`
	Start        string `json:"start" command:"name:start"`
	Bond         string `json:"bond" command:"name:bond"`
}

// Validate is a method that validates the dto
func (c *ContractAddIn) Validate(repo port.Repository) error {
	if c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetDomain is a method that returns a domain representation of the contract dto
func (c *ContractAddIn) GetDomain() []port.Domain {
	if c.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		c.Date = time.Now().Format(pkg.DateFormat)
	}
	return []port.Domain{
		domain.NewContract(c.ID, c.Date, c.ClientID, c.ServiceID, c.RecurrenceID, c.PriceID, c.BillingType, c.DueDay, c.Start, "", c.Bond),
	}
}

// GetOut is a method that returns the output dto
func (c *ContractAddIn) GetOut() port.DTOOut {
	return &ContractAddOut{}
}

// GetDTO is a method that returns the output dto
func (c *ContractAddOut) GetDTO(domainIn interface{}) interface{} {
	slices := domainIn.([]interface{})
	contract := slices[0].(*domain.Contract)
	dto := &ContractAddOut{}
	dto.ID = contract.ID
	dto.Date = contract.Date.Format(pkg.DateFormat)
	dto.ClientID = contract.ClientID
	dto.ServiceID = contract.ServiceID
	dto.RecurrenceID = contract.RecurrenceID
	dto.PriceID = contract.PriceID
	dto.BillingType = contract.BillingType
	dto.DueDay = strconv.FormatInt(contract.DueDay, 10)
	dto.Start = contract.Start.Format(pkg.DateFormat)
	dto.Bond = *contract.Bond
	return dto
}

// isEmpty is a method that checks if the dto is empty
func (c *ContractAddIn) isEmpty() bool {
	return c.ID == "" || c.Date == "" || c.ClientID == "" || c.ServiceID == "" || c.RecurrenceID == "" || 
	       c.PriceID == "" || c.BillingType == "" || c.DueDay == "" || c.Start == "" || c.Bond == ""
}
