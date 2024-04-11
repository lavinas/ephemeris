package domain

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Cycle represents the cycle entity

var (
	Cycles = map[string]string{
		"once":  "once",
		"day":   "day",
		"week":  "week",
		"month": "month",
		"year":  "year",
	}
)

// Recurrence represents the recurrence entity
type Recurrence struct {
	ID     string    `gorm:"type:varchar(25); primaryKey"`
	Date   time.Time `gorm:"type:datetime; not null; index"`
	Name   string    `gorm:"type:varchar(100); not null; index"`
	Cycle  string    `gorm:"type:varchar(20); not null; index"`
	Length *int64    `gorm:"type:numeric(10); null; index"`
	Limits *int64    `gorm:"type:numeric(10); null; index"`
}

// NewRecurrence is a function that creates a new recurrence
func NewRecurrence(id string, date string, name string, cycle string, length string, limit string) *Recurrence {
	date = strings.TrimSpace(date)
	local, _ := time.LoadLocation(pkg.Location)
	fdate := time.Time{}
	if date != "" {
		var err error
		if fdate, err = time.ParseInLocation(pkg.DateFormat, date, local); err != nil {
			fdate = time.Time{}
		}
	}
	var flen *int64 = nil
	if len, _ := strconv.ParseInt(length, 10, 64); len > 0 {
		flen = &len
	}
	var flim *int64 = nil
	if lim, _ := strconv.ParseInt(limit, 10, 64); lim > 0 {
		flim = &lim
	}
	return &Recurrence{
		ID:     id,
		Date:   fdate,
		Name:   name,
		Cycle:  cycle,
		Length: flen,
		Limits: flim,
	}
}

// Format is a method that formats the recurrence
func (r *Recurrence) Format(repo port.Repository, args ...string) error {
	filled := slices.Contains(args, "filled")
	noduplicity := slices.Contains(args, "noduplicity")
	msg := ""
	if err := r.formatID(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := r.formatDate(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := r.formatName(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := r.formatCycle(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := r.formatLength(filled); err != nil {
		msg += err.Error() + " | "
	}
	if err := r.formatLimit(); err != nil {
		msg += err.Error() + " | "
	}
	if err := r.validateDuplicity(repo, noduplicity); err != nil {
		msg += err.Error() + " | "
	}
	if msg != "" {
		return fmt.Errorf(msg[:len(msg)-3])
	}
	return nil
}

// Exists is a method that checks if the recurrence exists
func (r *Recurrence) Exists(repo port.Repository) (bool, error) {
	return repo.Get(&Recurrence{}, r.ID)
}

// GetID is a method that returns the id of the recurrence
func (r *Recurrence) GetID() string {
	return r.ID
}

// Get is a method that returns the recurrence
func (r *Recurrence) Get() port.Domain {
	return r
}

// GetEmpty is a method that returns an empty recurrence with just id
func (r *Recurrence) GetEmpty() port.Domain {
	return &Recurrence{}
}

// TableName returns the table name for database
func (r *Recurrence) TableName() string {
	return "recurrence"
}

// formatID is a method that formats the recurrence id
func (r *Recurrence) formatID(filled bool) error {
	r.ID = r.formatString(r.ID)
	if r.ID == "" {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrEmptyID)
	}
	if len(r.ID) > 25 {
		return fmt.Errorf(pkg.ErrLongID)
	}
	if len(strings.Split(r.ID, " ")) > 1 {
		return fmt.Errorf(pkg.ErrInvalidID)
	}
	return nil
}

// formatDate is a method that formats the recurrence date
func (r *Recurrence) formatDate(filled bool) error {
	if r.Date.IsZero() {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrInvalidDateFormat)
	}
	return nil
}

// formatName is a method that formats the recurrence name
func (r *Recurrence) formatName(filled bool) error {
	r.Name = r.formatString(r.Name)
	if r.Name == "" {
		if filled {
			return nil
		}
		return fmt.Errorf(pkg.ErrEmptyName)
	}
	if len(r.Name) > 100 {
		return fmt.Errorf(pkg.ErrLongName)
	}
	return nil
}

// formatCycle is a method that formats the recurrence cycle
func (r *Recurrence) formatCycle(filled bool) error {
	r.Cycle = r.formatString(r.Cycle)
	if r.Cycle == "" {
		if filled {
			return nil
		}
		cycles := ""
		for k := range Cycles {
			cycles += k + ", "
		}
		return fmt.Errorf(pkg.ErrEmptyCycle, cycles[:len(cycles)-2])
	}
	if len(r.Cycle) > 20 {
		return fmt.Errorf(pkg.ErrLongCycle)
	}
	if _, ok := Cycles[r.Cycle]; !ok {
		cycles := ""
		for k := range Cycles {
			cycles += k + ", "
		}
		return fmt.Errorf(pkg.ErrInvalidCycle, cycles[:len(cycles)-2])
	}
	return nil
}

// formatQuantity is a method that formats the recurrence quantity
func (r *Recurrence) formatLength(filled bool) error {
	if filled && r.Length == nil {
		return nil
	}
	if r.Cycle == "once" && r.Length != nil {
		return fmt.Errorf(pkg.ErrZeroLen)
	}
	if r.Cycle != "once" && r.Length == nil {
		return fmt.Errorf(pkg.ErrEmptyLen)
	}
	return nil
}

// formatLimit is a method that formats the recurrence limit
func (r *Recurrence) formatLimit() error {
	if r.Limits == nil {
		return nil
	}
	if r.Cycle == "once" && r.Limits != nil {
		return fmt.Errorf(pkg.ErrZeroLimit)
	}
	return nil
}

// formatString is a method that formats a string
func (r *Recurrence) formatString(str string) string {
	str = strings.TrimSpace(str)
	space := regexp.MustCompile(`\s+`)
	str = space.ReplaceAllString(str, " ")
	return str
}

// validateDuplicity is a method that validates the duplicity of a recurrence
func (r *Recurrence) validateDuplicity(repo port.Repository, noduplicity bool) error {
	if noduplicity {
		return nil
	}
	ok, err := repo.Get(&Recurrence{}, r.ID)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf(pkg.ErrAlreadyExists, r.ID)
	}
	return nil
}
