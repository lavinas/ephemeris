package usecase

import (
	"errors"
	"reflect"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Crud is a struct that groups the crud usecase
type Crud struct {
	Repo port.Repository
	Log  port.Logger
	Out  []port.DTOOut
}

// NewAdd is a function that returns a new Add struct
func NewCrud(repo port.Repository, log port.Logger) *Crud {
	return &Crud{
		Repo: repo,
		Log:  log,
		Out:  nil,
	}
}

// SetRepo is a method that sets the repository
func (c *Crud) SetRepo(repo port.Repository) {
	c.Repo = repo
}

// SetLog is a method that sets the logger
func (c *Crud) SetLog(log port.Logger) {
	c.Log = log
}

// Run is a method that runs the use case
func (c *Crud) Run(dto interface{}) error {
	in := dto.(port.DTOIn)
	switch in.GetCommand() {
	case "add":
		return c.Add(in)
	case "get":
		return c.Get(in)
	case "up":
		return c.Up(in)
	default:
		return c.error(pkg.ErrPrefCommandNotFound, pkg.ErrCommandNotFound)
	}
}

// Interface is a method that returns the output dto as an interface
func (c *Crud) Interface() interface{} {
	return c.Out
}

// String is a method that returns a string representation of the output dto
func (c *Crud) String() string {
	if c.Out == nil {
		return ""
	}
	return pkg.NewCommands().Marshal(c.Out, "trim", "nokeys")
}

// Add is a method that add a dto to the repository
func (c *Crud) Add(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error())
	}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	defer c.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	for _, domain := range domains {
		if err := domain.Format(c.Repo); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error())
		}
		if err := c.Repo.Add(domain); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		result = append(result, c.sliceOf(domain))
	}
	if err := c.Repo.Commit(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}

// Get is a method that gets a dto from the repository
func (c *Crud) Get(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error())
	}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	defer c.Repo.Rollback()
	domains := in.GetDomain()
	result := []interface{}{}
	for _, domain := range domains {
		if err := domain.Format(c.Repo, "filled", "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error())
		}
		found, err := c.Repo.Find(domain)
		if err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		if found == nil {
			return c.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound)
		}
		result = append(result, found)
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}

// Up is a method that updates a dto in the repository
func (c *Crud) Up(dtoIn interface{}) error {
	in := dtoIn.(port.DTOIn)
	if err := in.Validate(c.Repo); err != nil {
		return c.error(pkg.ErrPrefBadRequest, err.Error())
	}
	domains := in.GetDomain()
	result := []interface{}{}
	if err := c.Repo.Begin(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	defer c.Repo.Rollback()
	for _, source := range domains {
		if err := source.Format(c.Repo, "filled", "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefBadRequest, err.Error())
		}
		target := source.GetEmpty()
		if f, err := c.Repo.Get(target, source.GetID()); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		} else if !f {
			return c.error(pkg.ErrPrefBadRequest, pkg.ErrUnfound)
		}
		if err := c.merge(source, target); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := target.Format(c.Repo, "noduplicity"); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		if err := c.Repo.Save(target); err != nil {
			return c.error(pkg.ErrPrefInternal, err.Error())
		}
		result = append(result, c.sliceOf(target))
	}
	if err := c.Repo.Commit(); err != nil {
		return c.error(pkg.ErrPrefInternal, err.Error())
	}
	out := in.GetOut()
	c.Out = out.GetDTO(result)
	return nil
}

// sliceOf is a method that returns a slice of a struct
func (c *Crud) sliceOf(in interface{}) interface{} {
	ret := reflect.New(reflect.SliceOf(reflect.TypeOf(in).Elem()))
	val := ret.Elem()
	val.Set(reflect.Append(val, reflect.ValueOf(in).Elem()))
	return ret.Interface()
}

// merge is a method that merges two structs
func (c *Crud) merge(source interface{}, target interface{}) error {
	if reflect.TypeOf(source) != reflect.TypeOf(target) {
		return c.error(pkg.ErrPrefInternal, pkg.ErrInvalidTypeOnMerge)
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
func (c *Crud) error(prefix string, err string) error {
	err = prefix + ": " + err
	c.Log.Println(err)
	return errors.New(err)
}
