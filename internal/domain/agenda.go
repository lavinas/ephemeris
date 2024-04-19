package domain

import (
	"strings"
	"time"

	"github.com/lavinas/ephemeris/pkg"
)

// AgendaType represents the agenda type entity

// Agenda represents the agenda entity
type Agenda struct {
	ID           string     `gorm:"type:varchar(25); primaryKey"`
	Date         time.Time  `gorm:"type:datetime; not null"`
	ContractID   string     `gorm:"type:varchar(25); not null; index"`
	Start        time.Time  `gorm:"type:datetime; not null"`
	End          time.Time  `gorm:"type:datetime; not null"`
	Kind         string     `gorm:"type:varchar(25); not null; index"`
	Status       string     `gorm:"type:varchar(25); not null; index"`
	Bond         *Agenda    `gorm:"foreignKey:ID"`
	BillingMonth *time.Time `gorm:"type:numeric(20)"`
}

// NewAgenda creates a new agenda domain entity
func NewAgenda(id, date, contractID, start, end, kind, status, bond, billing string) *Agenda {
	agenda := &Agenda{}
	agenda.ID = id
	local, _ := time.LoadLocation(pkg.Location)
	agenda.Date, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(date), local)
	agenda.ContractID = contractID
	agenda.Start, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(start), local)
	agenda.End, _ = time.ParseInLocation(pkg.DateFormat, strings.TrimSpace(end), local)
	agenda.Kind = kind
	agenda.Status = status
	if bond != "" {
		agenda.Bond = &Agenda{ID: bond}
	}
	if billing != "" {
		mont, err := time.ParseInLocation(pkg.MonthFormat, billing, local)
		if err != nil {
			mont, _ = time.ParseInLocation(pkg.DateFormat, billing, local)
		}
		agenda.BillingMonth = &mont
	}
	return agenda
}
