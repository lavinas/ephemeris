package usecase

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Add is a struct that groups the Add usecase
type Add struct {
	Repo port.Repository
	Log  port.Logger
	Out  port.DTOOut
}

// NewAdd is a function that returns a new Add struct
func NewAdd(repo port.Repository, log port.Logger) *Add {
	return &Add{
		Repo: repo,
		Log:  log,
		Out:  nil,
	}
}

// SetRepo is a method that sets the repository
func (u *Add) SetRepo(repo port.Repository) {
	u.Repo = repo
}

// SetLog is a method that sets the logger
func (u *Add) SetLog(log port.Logger) {
	u.Log = log
}

// Add is a method that add a dto to the repository
func (u *Add) Run(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	for _, domain := range domains {
		if err := domain.Format(u.Repo); err != nil {
			return u.error(pkg.ErrPrefBadRequest, err.Error())
		}
		if err := u.Repo.Add(domain); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
		result = append(result, domain)
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	out := in.GetOut()
	u.Out = out.GetDTO(result).(port.DTOOut)
	return nil
}

// Interface is a method that returns the output dto as an interface
func (u *Add) Interface() interface{} {
	return u.Out
}

// String is a method that returns a string representation of the output dto
func (y *Add) String() string {
	if y.Out == nil {
		return ""
	}
	return pkg.NewCommands().Marshal(y.Out, "trim")
}

// error is a function that logs an error and returns it
func (u *Add) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
