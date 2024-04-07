package usecase

import (
	"errors"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Get is a struct that groups the Get usecase
type Get struct {
	Repo port.Repository
	Log  port.Logger
	Out  interface{}
}

// NewGet is a function that returns a new Get struct
func NewGet(repo port.Repository, log port.Logger) *Get {
	return &Get{
		Repo: repo,
		Log:  log,
		Out:  nil,
	}
}

// SetRepo is a method that sets the repository
func (u *Get) SetRepo(repo port.Repository) {
	u.Repo = repo
}

// SetLog is a method that sets the logger
func (u *Get) SetLog(log port.Logger) {
	u.Log = log
}

// Get is a method that gets a dto from the repository
func (u *Get) Run(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(port.ErrPrefBadRequest, err.Error())
	}
	if err := u.Repo.Begin(); err != nil {
		return u.error(port.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	for _, domain := range domains {
		if err := domain.Format(u.Repo, "filled", "noduplicity"); err != nil {
			return u.error(port.ErrPrefBadRequest, err.Error())
		}
		found, err := u.Repo.Find(domain)
		if err != nil {
			return u.error(port.ErrPrefInternal, err.Error())
		}
		if found == nil {
			return u.error(port.ErrPrefBadRequest, port.ErrUnfound)
		}
		result = append(result, found)
	}
	out := in.GetOut()
	u.Out = out.GetDTO(result)
	return nil
}

// String is a method that returns the output dto as a string
func (u *Get) String() string {
	if u.Out == nil {
		return ""
	}
	return pkg.NewCommands().Marshal(u.Out, "trim")
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
