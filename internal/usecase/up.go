package usecase

import (
	"errors"
	"reflect"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

type Up struct {
	Repo port.Repository
	Log  port.Logger
	Out  port.DTOOut
}

// NewUp is a function that returns a new Up struct
func NewUp(repo port.Repository, log port.Logger) *Up {
	return &Up{
		Repo: repo,
		Log:  log,
	}
}

// SetRepo is a method that sets the repository
func (u *Up) SetRepo(repo port.Repository) {
	u.Repo = repo
}

// SetLog is a method that sets the logger
func (u *Up) SetLog(log port.Logger) {
	u.Log = log
}

// Up is a method that updates a dto in the repository
func (u *Up) Run(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	domains := in.GetDomain()
	result := []interface{}{}
	if err := u.Repo.Begin(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	defer u.Repo.Rollback()
	for _, source := range domains {
		if err := source.Format(u.Repo, "filled", "noduplicity"); err != nil {
			return u.error(pkg.ErrPrefBadRequest, err.Error())
		}
		target := source.GetEmpty()
		if f, err := u.Repo.Get(target, source.GetID()); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		} else if !f {
			return u.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound)
		}
		if err := u.merge(source, target); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := u.Repo.Save(target); err != nil {
			return u.error(pkg.ErrPrefInternal, err.Error())
		}
		result = append(result, target)
	}
	if err := u.Repo.Commit(); err != nil {
		return u.error(pkg.ErrPrefInternal, err.Error())
	}
	out := in.GetOut()
	u.Out = out.GetDTO(result).(port.DTOOut)
	return nil
}

// String is a method that returns the output dto as a string
func (u *Up) String() string {
	return pkg.NewCommands().Marshal(u.Out, "trim")
}

// Interface is a method that returns the output dto as an interface
func (u *Up) Interface() interface{} {
	return u.Out
}

// merge is a method that merges two structs
func (u *Up) merge(source interface{}, target interface{}) error {
	if reflect.TypeOf(source) != reflect.TypeOf(target) {
		return u.error(pkg.ErrPrefInternal, pkg.ErrInvalidTypeOnMerge)
	}
	s := reflect.ValueOf(source).Elem()
	t := reflect.ValueOf(target).Elem()
	for i := 0; i < s.NumField(); i++ {
		if s.Field(i).Interface() != reflect.Zero(s.Field(i).Type()).Interface() {
			t.Field(i).Set(s.Field(i))
		}
	}
	return nil
}

// error is a function that logs an error and returns it
func (u *Up) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
