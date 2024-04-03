package usecase

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Get is a struct that groups the Get usecase
type Get struct {
	Repo port.Repository
	Log  port.Logger
	Out  []port.DTOOut
}

// NewGet is a function that returns a new Get struct
func NewGet(repo port.Repository, log port.Logger) *Get {
	return &Get{
		Repo: repo,
		Log:  log,
		Out:  nil,
	}
}

// Get is a method that gets a dto from the repository
func (u *Get) Run(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return err
	}
	domain := in.GetDomain()
	if err := domain.Format("filled"); err != nil {
		err := u.error(port.ErrPrefBadRequest, err.Error())
		return err
	}
	found, err := u.Repo.Find(domain)
	if err != nil {
		err := u.error(port.ErrPrefInternal, err.Error())
		return err
	}
	out := dto.ClientGetOut{}
	u.Out = out.GetDTO(found.([]port.Domain)).([]port.DTOOut)
	return nil
}

// Dto is a method that returns the output dto
func (u *Get) Dto() []port.DTOOut {
	return u.Out
}

// String is a method that returns the output dto as a string
func (u *Get) String() string {
	if u.Out == nil {
		return ""
	}
	return pkg.NewCommands().Marshal(u.Out)
}

// Interface is a method that returns the output dto as an interface
func (u *Get) Interface() interface{} {
	return u.Out
}

// error is a function that logs an error and returns it
func (u *Get) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
