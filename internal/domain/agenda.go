package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"context"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	kindAgenda = []string{
		pkg.AgendaKindRegular,
		pkg.AgendaKindRescheduled,
		pkg.AgendaKindExtra,
	}
	statusAgenda = []string{
		pkg.AgendaStatusOpenned,
		pkg.AgendaStatusDone,
		pkg.AgendaStatusSaved,
		pkg.AgendaStatusMissed,
		pkg.AgendaStatusCanceled,
		pkg.AgendaStatusLocked,
	}
)

// Agenda represents the agenda entity
type Agenda struct {
	ID           string     `gorm:"type:varchar(150); primaryKey"`
	Date         time.Time  `gorm:"type:datetime; not null"`
	ClientID     string     `gorm:"type:varchar(50); not null; index"`
	ServiceID    string     `gorm:"type:varchar(50); not null; index"`
	ContractID   *string    `gorm:"type:varchar(50); null; index"`
	Start        time.Time  `gorm:"type:datetime; not null"`
	End          time.Time  `gorm:"type:datetime; not null"`
	Price        *float64   `gorm:"type:decimal(10,2)"`
	Kind         string     `gorm:"type:varchar(50); not null; index"`
	Status       string     `gorm:"type:varchar(50); not null; index"`
	Bond         *string    `gorm:"type:varchar(50)"`
	BillingMonth *time.Time `gorm:"type:datetime"`
	Locked       *time.Time `gorm:"type:datetime;null; index"`
}

// NewAgenda creates a new agenda domain entity
func NewAgenda(id, date, clientID, serviceID, contractID, start, end, price, kind, status, bond, billing string) *Agenda {
	agenda := &Agenda{}
	agenda.ID = id
	local, _ := time.LoadLocation(pkg.Location)
	agenda.Date, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(date), local)
	agenda.ServiceID = serviceID
	agenda.ClientID = clientID
	if contractID != "" {
		agenda.ContractID = &contractID
	}
	agenda.Start, _ = time.ParseInLocation(pkg.DateTimeFormat, strings.TrimSpace(start), local)
	agenda.End, _ = time.ParseInLocation(pkg.DateTimeFormat, strings.TrimSpace(end), local)
	agenda.Kind = kind
	agenda.Status = status
	if bond != "" {
		agenda.Bond = &bond
	}
	mont, err := time.ParseInLocation(pkg.MonthFormat, billing, local)
	if err == nil && !mont.IsZero() {
		agenda.BillingMonth = &mont
	} else {
		mont, err = time.ParseInLocation(pkg.DateFormat, billing, local)
		if err == nil && !mont.IsZero() {
			agenda.BillingMonth = &mont
		}
	}
	if p, err := strconv.ParseFloat(price, 64); err == nil {
		agenda.Price = &p
	}
	return agenda
}

// Format formats the agenda
func (a *Agenda) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := a.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatContractID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatClientID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatServiceID(repo, filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatStart(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatEnd(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatKind(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatPrice(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatStatus(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatBond(repo); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.formatBillingMonth(); err != nil {
		msg += err.Error() + " | "
	}
	if err := a.validateDuplicity(repo, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg == "" {
		return nil
	}
	return errors.New(msg[:len(msg)-3])
}

// Exists is a function that checks if a agenda exists
func (a *Agenda) Load(repo port.Repository) (bool, error) {
	return repo.Get(a, a.ID, "")
}

// LoadRange loads agenda slices from a interval of dates
func (a *Agenda) LoadRange(repo port.Repository, start, end time.Time, status []string) ([]*Agenda, error) {
	extras := a.loadRangeExtras(start, end, status)
	agendas, _, err := repo.Find(a, 0, "", extras...)
	if err != nil {
		return nil, err
	}
	if agendas == nil {
		return nil, nil
	}
	ag := agendas.(*[]Agenda)
	ret := make([]*Agenda, 0)
	for _, a := range *ag {
		ret = append(ret, &a)
	}
	return ret, nil
}

// GetID is a method that returns the id of the client
func (a *Agenda) GetID() string {
	return a.ID
}

// Get is a method that returns the client
func (a *Agenda) Get() port.Domain {
	return a
}

// GetEmpty is a method that returns an empty client with just id
func (a *Agenda) GetEmpty() port.Domain {
	return &Agenda{}
}

// Lock is a method that locks the contract
func (a *Agenda) Lock(repo port.Repository, timeout int) error {
	if a.IsLocked(repo, timeout) {
		return errors.New(pkg.ErrAgendaLocked)
	}
	x := time.Now()
	a.Locked = &x 
	if err := repo.Save(a, ""); err != nil {
		return err
	}
	return nil
}

// IsLocked is a method that checks if the contract is locked
func (a *Agenda) IsLocked(repo port.Repository, timeout int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	for {
		if a.Locked == nil {
			return false
		}
		select {
		case <-ctx.Done():
			return true
		default:
			time.Sleep(1 * time.Second)
			ag, error := a.getHotAgenda(repo)
			if error != nil {
				return true
			}
			a.Locked = ag.Locked
		}
	}
}

// getHot gets the agenda out of default transaction
func (a *Agenda) getHotAgenda(repo port.Repository) (*Agenda, error) {
	tx := repo.NewTransaction()
	if err := repo.Begin(tx); err != nil {
		return nil, err
	}
	defer repo.Rollback(tx)
	x := &Agenda{ID: a.ID}
	if ok, err := repo.Get(x, a.ID, tx); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New(pkg.ErrAgendaNotFound)
	}
	return x, nil
}

// Unlock is a method that unlocks the contract
func (a *Agenda) Unlock(repo port.Repository) error {
	a.Locked = nil
	if err := repo.Save(a, ""); err != nil {
		return err
	}
	return nil
}

// TableName returns the table name for database
func (a *Agenda) TableName() string {
	return "agenda"
}

// loadRangeExtras is a method that monts the load range extras
func (a *Agenda) loadRangeExtras(start, end time.Time, status []string) []interface{} {
	extras := []interface{}{}
	if !start.IsZero() {
		st := fmt.Sprintf("Start >= '%d-%02d-%02d %02d:%02d:%02d'", start.Year(), start.Month(), start.Day(),
			start.Hour(), start.Minute(), start.Second())
		extras = append(extras, st)
	}
	if !end.IsZero() {
		ed := fmt.Sprintf("Start <= '%d-%02d-%02d %02d:%02d:%02d'", end.Year(), end.Month(), end.Day(),
			end.Hour(), end.Minute(), end.Second())
		extras = append(extras, ed)
	}
	q := ("(")
	if status != nil {
		for _, s := range status {
			if !slices.Contains(statusAgenda, s) {
				return nil
			}
			q += fmt.Sprintf("Status = '%s' OR ", s)
		}
		if q != "(" {
			q = q[:len(q)-4] + ")"
			extras = append(extras, q)
		}
	}
	return extras
}

// formatID is a method that formats the id of the contract
func (c *Agenda) formatID(filled bool) error {
	id := c.formatString(c.ID)
	if id == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyID)
	}
	if len(id) > 150 {
		return errors.New(pkg.ErrLongID150)
	}
	if len(strings.Split(id, " ")) > 1 {
		return errors.New(pkg.ErrInvalidID)
	}
	c.ID = strings.ToLower(id)
	return nil
}

// formatDate is a method that formats the date of the contract
func (c *Agenda) formatDate(filled bool) error {
	if c.Date.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidDateFormat, pkg.DateFormat)
	}
	return nil
}

// formatClientID is a method that formats the client id
func (c *Agenda) formatClientID(repo port.Repository, filled bool) error {
	clientID := c.formatString(c.ClientID)
	if clientID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyClientID)
	}
	client := &Client{ID: clientID}
	if exists, err := client.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrClientNotFound)
	}
	return nil
}

// formarServiceID is a method that formats the service id
func (c *Agenda) formatServiceID(repo port.Repository, filled bool) error {
	serviceID := c.formatString(c.ServiceID)
	if serviceID == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyServiceID)
	}
	service := &Service{ID: serviceID}
	if exists, err := service.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrServiceNotFound)
	}
	return nil
}

// formatContractID is a method that formats the contract id
func (c *Agenda) formatContractID(repo port.Repository, filled bool) error {
	if c.ContractID == nil {
		if c.Price != nil {
			if filled {
				return nil
			}
			return errors.New(pkg.ErrContractIDNotProvided)
		}
		return nil
	}
	contract := &Contract{ID: *c.ContractID}
	if exists, err := contract.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrContractNotFound)
	}
	if c.ClientID == "" {
		c.ClientID = contract.ClientID
	} else if contract.ClientID != c.ClientID {
		return errors.New(pkg.ErrContractClientMismatch)
	}
	return nil
}

// formatStart is a method that formats the start date of the agenda
func (c *Agenda) formatStart(filled bool) error {
	if filled {
		return nil
	}
	if c.Start.IsZero() {
		return fmt.Errorf(pkg.ErrInvalidStartDate, pkg.DateTimeFormat)
	}
	return nil
}

// formatEnd is a method that formats the end date of the agenda
func (c *Agenda) formatEnd(filled bool) error {
	if c.End.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidEndDate, pkg.DateTimeFormat)
	}
	if c.Start.After(c.End) {
		return errors.New(pkg.ErrStartAfterEndDate)
	}
	return nil
}

// formatKind is a method that formats the kind of the agenda
func (c *Agenda) formatKind(filled bool) error {
	kind := c.formatString(c.Kind)
	if kind == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyKind)
	}
	if !slices.Contains(kindAgenda, kind) {
		return fmt.Errorf(pkg.ErrInvalidKind, strings.Join(kindAgenda, ", "))
	}
	c.Kind = kind
	return nil
}

// formatPrice is a method that formats the price of the agenda
func (c *Agenda) formatPrice(filled bool) error {
	if c.Price == nil {
		if c.ContractID == nil {
			if filled {
				return nil
			}
			return errors.New(pkg.ErrPriceIDNotProvided)
		}
		return nil
	}
	if *c.Price < 0 {
		return errors.New(pkg.ErrInvalidPackPrice)
	}
	return nil
}

// formatStatus is a method that formats the status of the agenda
func (c *Agenda) formatStatus(filled bool) error {
	status := c.formatString(c.Status)
	if status == "" {
		if filled {
			return nil
		}
		return errors.New(pkg.ErrEmptyStatus)
	}
	if !slices.Contains(statusAgenda, status) {
		return fmt.Errorf(pkg.ErrInvalidStatus, strings.Join(cycles, ", "))
	}
	c.Status = status
	return nil
}

// formatBond is a method that formats the bond of the agenda
func (c *Agenda) formatBond(repo port.Repository) error {
	if c.Bond == nil {
		return nil
	}
	bond := &Agenda{ID: *c.Bond}
	if exists, err := bond.Load(repo); err != nil {
		return err
	} else if !exists {
		return errors.New(pkg.ErrBondNotFound)
	}
	return nil
}

// formatBillingMonth is a method that formats the billing month of the agenda
func (c *Agenda) formatBillingMonth() error {
	if c.BillingMonth == nil {
		return nil
	}
	if c.BillingMonth.IsZero() {
		return fmt.Errorf(pkg.ErrInvalidBillingMonth, pkg.MonthFormat)
	}
	return nil
}

// validateDuplicity is a method that validates the duplicity of a client
func (c *Agenda) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Agenda{}, c.ID, "")
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, c.ID)
	}
	return nil
}

// formatString is a method that formats a string
func (c *Agenda) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}
