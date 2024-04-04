package usecase

import (
	"errors"
	"fmt"

	"github.com/lavinas/ephemeris/internal/dto"
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
	if err := in.Validate(); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return err
	}
	domains := in.GetDomain()
	result := []interface{}{}
	for _, domain := range domains {
		if err := domain.Format(); err != nil {
			err := u.error(port.ErrPrefBadRequest, err.Error())
			return err
		}
		if f, err := u.Repo.Get(domain, domain.GetID()); err != nil {
			err := u.error(port.ErrPrefInternal, err.Error())
			return err
		} else if f {
			err := u.error(port.ErrPrefConflict, fmt.Sprintf(port.ErrAlreadyExists, domain.GetID()))
			return err
		}
		if err := u.Repo.Add(domain); err != nil {
			err := u.error(port.ErrPrefInternal, err.Error())
			return err
		}
		result = append(result, domain)
	}
	out := dto.ClientAddOut{}
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
	return pkg.NewCommands().Marshal(y.Out)
}

// error is a function that logs an error and returns it
func (u *Add) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
