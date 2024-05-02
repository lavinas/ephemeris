package dto

import (
	"errors"
	"strconv"
	"time"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ContractCrud represents the dto for getting a contract
type ContractCrud struct {
	Object      string `json:"-" command:"name:contract;key;pos:2-"`
	Action      string `json:"-" command:"name:add,get,up;key;pos:2-"`
	ID          string `json:"id" command:"name:id;pos:3+"`
	Date        string `json:"date" command:"name:date;pos:3+"`
	ClientID    string `json:"client" command:"name:client;pos:3+"`
	SponsorID   string `json:"sponsor" command:"name:sponsor;pos:3+"`
	PackageID   string `json:"package" command:"name:package;pos:3+"`
	BillingType string `json:"billing" command:"name:billing;pos:3+"`
	DueDay      string `json:"due" command:"name:due;pos:3+"`
	Start       string `json:"start" command:"name:start;pos:3+"`
	End         string `json:"end" command:"name:end;pos:3+"`
	Bond        string `json:"bond" command:"name:bond;pos:3+"`
	Locked      string `json:"locked" command:"name:locked;pos:3+"`
}

// Validate is a method that validates the dto
func (c *ContractCrud) Validate(repo port.Repository) error {
	if c.Action != "get" && c.isEmpty() {
		return errors.New(pkg.ErrParamsNotInformed)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (c *ContractCrud) GetCommand() string {
	return c.Action
}

// GetDomain is a method that returns the domain of the dto
func (c *ContractCrud) GetDomain() []port.Domain {
	time.Local, _ = time.LoadLocation(pkg.Location)
	if c.Action == "add" && c.Date == "" {
		c.Date = time.Now().Format(pkg.DateFormat)
	}
	if c.Action == "add" && c.Start == "" {
		c.Start = time.Now().Format(pkg.DateTimeFormat)
	}
	if c.Action == "add" && c.DueDay == "" && c.BillingType != pkg.BillingTypePerSession {
		c.DueDay = pkg.DefaultDueDay
	}
	if c.Action == "add" && c.BillingType == "" {
		c.BillingType = pkg.DefaultBillingType
	}
	return []port.Domain{
		domain.NewContract(c.ID, c.Date, c.ClientID, c.SponsorID, c.PackageID, c.BillingType, c.DueDay, c.Start, c.End, c.Bond),
	}
}

// GetOut is a method that returns the dto out
func (c *ContractCrud) GetOut() port.DTOOut {
	return &ContractCrud{}
}

// GetDTO is a method that returns the dto
func (c *ContractCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	contracts := slices[0].(*[]domain.Contract)
	for _, c := range *contracts {
		sponsor := ""
		if c.SponsorID != nil {
			sponsor = *c.SponsorID
		}
		due := ""
		if c.DueDay != nil {
			due = strconv.FormatInt(*c.DueDay, 10)
		}
		end := ""
		if c.End != nil {
			end = c.End.Format(pkg.DateFormat)
		}
		bond := ""
		if c.Bond != nil {
			bond = *c.Bond
		}
		locked := ""
		if c.Locked != nil && *c.Locked {
			locked = "******"
		}
		ret = append(ret, &ContractCrud{
			ID:          c.ID,
			Date:        c.Date.Format(pkg.DateFormat),
			ClientID:    c.ClientID,
			SponsorID:   sponsor,
			PackageID:   c.PackageID,
			BillingType: c.BillingType,
			DueDay:      due,
			Start:       c.Start.Format(pkg.DateTimeFormat),
			End:         end,
			Bond:        bond,
			Locked:      locked,
		})
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

// isEmpty is a method that checks if the dto is empty
func (c *ContractCrud) isEmpty() bool {
	return c.ID == "" && c.Date == "" && c.ClientID == "" && c.SponsorID == "" && c.PackageID == "" &&
		c.BillingType == "" && c.DueDay == "" && c.Start == "" && c.End == "" && c.Bond == ""
}
