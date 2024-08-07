package usecase

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	runMap = map[string]func(*Usecase, interface{}) error{
		"add":     (*Usecase).Add,
		"get":     (*Usecase).Get,
		"up":      (*Usecase).Up,
		"make":    (*Usecase).AgendaMake,
		"tie":     (*Usecase).SessionTie,
		"untie":   (*Usecase).SessionTie,
		"confirm": (*Usecase).SessionTie,
		"force":   (*Usecase).SessionForce,
	}
)

// Usecase is a struct that groups the crud usecase
type Usecase struct {
	Repo    port.Repository
	Log     port.Logger
	Out     []port.DTOOut
	Limited bool
}

// NewAdd is a function that returns a new Add struct
func NewUsecase(repo port.Repository, log port.Logger) *Usecase {
	return &Usecase{
		Repo:    repo,
		Log:     log,
		Out:     nil,
		Limited: false,
	}
}

// Run is a method that runs the use case
func (c *Usecase) Run(dto interface{}) error {
	in := dto.(port.DTOIn)
	cmd := in.GetCommand()
	if runMap[cmd] == nil {
		return c.error(pkg.ErrPrefBadRequest, pkg.ErrCommandNotFound, 0, 0)
	}
	return runMap[cmd](c, dto)
}

// Interface is a method that returns the output dto as an interface
//
//	and a boolean that indicates if the output was limited
func (c *Usecase) Interface() (interface{}, bool) {
	return c.Out, c.Limited
}

// String is a method that returns a string representation of the output dto
func (c *Usecase) String() string {
	if c.Out == nil {
		return ""
	}
	if c.Limited {
		return pkg.NewCommands().Marshal(c.Out, "trim", "nokeys", "more", "counter")
	}
	return pkg.NewCommands().Marshal(c.Out, "trim", "nokeys", "counter")
}

// sliceOf is a method that returns a slice of a struct
func (c *Usecase) sliceOf(in interface{}) interface{} {
	ret := reflect.New(reflect.SliceOf(reflect.TypeOf(in).Elem()))
	val := ret.Elem()
	val.Set(reflect.Append(val, reflect.ValueOf(in).Elem()))
	return ret.Interface()
}

// merge is a method that merges two structs
func (c *Usecase) merge(source interface{}, target interface{}) error {
	if reflect.TypeOf(source) != reflect.TypeOf(target) {
		return c.error(pkg.ErrPrefInternal, pkg.ErrInvalidTypeOnMerge, 0, 0)
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
func (c *Usecase) error(prefix string, err string, line int, lines int) error {
	err = prefix + ": " + err
	if lines > 1 {
		err = fmt.Sprintf("%s at line %d of %d", err, line, lines)
	}
	c.Log.Println(err)
	return errors.New(err)
}
