package dto

import (
	"errors"
	"strconv"
	"time"
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// ContractCrud represents the dto for getting a contract
type ContractCrud struct {
	Base
	Object      string `json:"-" command:"name:contract;key;pos:2-"`
	Action      string `json:"-" command:"name:add,get,up;key;pos:2-"`
	Sort        string `json:"sort" command:"name:sort;pos:3+"`
	Csv         string `json:"csv" command:"name:csv;pos:3+;" csv:"file"`
	ID          string `json:"id" command:"name:id;pos:3+;trans:id,string" csv:"id"`
	Date        string `json:"date" command:"name:date;pos:3+;trans:date,time" csv:"date"`
	ClientID    string `json:"client" command:"name:client;pos:3+;trans:client_id,string" csv:"client"`
	SponsorID   string `json:"sponsor" command:"name:sponsor;pos:3+;trans:sponsor_id,string" csv:"sponsor"`
	PackageID   string `json:"package" command:"name:package;pos:3+;trans:package_id,string" csv:"package"`
	BillingType string `json:"billing" command:"name:billing;pos:3+;trans:billing_type,string" csv:"billing"`
	DueDay      string `json:"due" command:"name:due;pos:3+;trans:due_day,int64" csv:"due"`
	Start       string `json:"start" command:"name:start;pos:3+;trans:start,time" csv:"start"`
	End         string `json:"end" command:"name:end;pos:3+;trans:end,time" csv:"end"`
	Bond        string `json:"bond" command:"name:bond;pos:3+;trans:bond,string" csv:"bond"`
	Locked      string `json:"locked" command:"name:locked;pos:3+;trans:locked,string" csv:"locked"`
}

// Validate is a method that validates the dto
func (c *ContractCrud) Validate() error {
	if c.Csv != "" && (c.ID != "" || c.Date != "" || c.ClientID != "" || c.SponsorID != "" || c.PackageID != "" ||
		c.BillingType != "" || c.DueDay != "" || c.Start != "" || c.End != "" || c.Bond != "" || c.Locked != "") {
		return errors.New(pkg.ErrCsvAndParams)
	}
	return nil
}

// GetCommand is a method that returns the command of the dto
func (c *ContractCrud) GetCommand() string {
	return c.Action
}

// GetDomain is a method that returns the domain of the dto
func (c *ContractCrud) GetDomain() []port.Domain {
	if c.Csv != "" {
		domains := []port.Domain{}
		contracts := []*ContractCrud{}
		c.ReadCSV(&contracts, c.Csv)
		for _, contract := range contracts {
			contract.Action = c.Action
			contract.Object = c.Object
			domains = append(domains, c.getDomain(contract))
		}
		return domains
	}
	return []port.Domain{c.getDomain(c)}
}

// GetOut is a method that returns the dto out
func (c *ContractCrud) GetOut() port.DTOOut {
	return c
}

// GetDTO is a method that returns the dto
func (c *ContractCrud) GetDTO(domainIn interface{}) []port.DTOOut {
	ret := []port.DTOOut{}
	slices := domainIn.([]interface{})
	for _, slice := range slices {
		contracts := slice.(*[]domain.Contract)
		for _, contract := range *contracts {
			sponsor := ""
			if contract.SponsorID != nil {
				sponsor = *contract.SponsorID
			}
			due := ""
			if contract.DueDay != nil {
				due = strconv.FormatInt(*contract.DueDay, 10)
			}
			end := ""
			if contract.End != nil {
				end = contract.End.Format(pkg.DateFormat)
			}
			bond := ""
			if contract.Bond != nil {
				bond = *contract.Bond
			}
			locked := ""
			if contract.Locked != nil && *contract.Locked {
				locked = "******"
			}
			ret = append(ret, &ContractCrud{
				ID:          contract.ID,
				Date:        contract.Date.Format(pkg.DateFormat),
				ClientID:    contract.ClientID,
				SponsorID:   sponsor,
				PackageID:   contract.PackageID,
				BillingType: contract.BillingType,
				DueDay:      due,
				Start:       contract.Start.Format(pkg.DateTimeFormat),
				End:         end,
				Bond:        bond,
				Locked:      locked,
			})
		}
	}
	pkg.NewCommands().Sort(ret, c.Sort)
	return ret
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (c *ContractCrud) GetInstructions(domain port.Domain) (port.Domain, []interface{}, error) {
	return c.getInstructions(c, domain)
}

// getDomain is a method that returns a string representation of the contract
func (c *ContractCrud) getDomain(one *ContractCrud) port.Domain {
	if one.Action == "add" && one.Date == "" {
		time.Local, _ = time.LoadLocation(pkg.Location)
		one.Date = time.Now().Format(pkg.DateFormat)
	}
	if one.Action == "add" && one.Start == "" {
		one.Start = time.Now().Format(pkg.DateTimeFormat)
	}
	if one.Action == "add" && one.DueDay == "" && one.BillingType != pkg.BillingTypePerSession {
		one.DueDay = pkg.DefaultDueDay
	}
	if one.Action == "add" && one.BillingType == "" {
		one.BillingType = pkg.DefaultBillingType
	}
	one.trim()
	return domain.NewContract(one.ID, one.Date, one.ClientID, one.SponsorID, one.PackageID, one.BillingType, one.DueDay, one.Start, one.End, one.Bond)
}

func (c *ContractCrud) trim() {
	c.ID = strings.TrimSpace(c.ID)
	c.Date = strings.TrimSpace(c.Date)
	c.ClientID = strings.TrimSpace(c.ClientID)
	c.SponsorID = strings.TrimSpace(c.SponsorID)
	c.PackageID = strings.TrimSpace(c.PackageID)
	c.BillingType = strings.TrimSpace(c.BillingType)
	c.DueDay = strings.TrimSpace(c.DueDay)
	c.Start = strings.TrimSpace(c.Start)
	c.End = strings.TrimSpace(c.End)
	c.Bond = strings.TrimSpace(c.Bond)
	c.Locked = strings.TrimSpace(c.Locked)
}
