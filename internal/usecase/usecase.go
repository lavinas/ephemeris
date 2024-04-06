package usecase

import (
	"errors"
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

var (
	dtos = map[interface{}]port.UseCase{
		&dto.ClientAddIn{}: &Add{},
		&dto.ClientGetIn{}: &Get{},
		&dto.ClientUpIn{}:  &Up{},
		&dto.LiableAddIn{}: &Add{},
	}
)

// Usecase is a struct that groups all usecases of the application
type CommandUsecase struct {
	Repo port.Repository
	Log  port.Logger
}

// UseCase is a function that returns a new UseCase struct
func NewCommandUsecase(repo port.Repository, log port.Logger) *CommandUsecase {
	repo.Migrate(domain.GetDomain())
	return &CommandUsecase{
		Repo: repo,
		Log:  log,
	}
}

// Run is a method that runs a command
func (u *CommandUsecase) Run(line string) string {
	u.Log.Println("Command: " + line)
	line = strings.ToLower(line)
	cmd := pkg.Commands{}
	inter := u.init()
	dtoIn, err := cmd.FindOne(line, inter)
	if err != nil {
		return u.error(port.ErrPrefCommandNotFound, err.Error()).Error()
	}
	if err := cmd.Unmarshal(line, dtoIn); err != nil {
		return u.error(port.ErrPrefBadRequest, err.Error()).Error()
	}
	dtoOut := dtos[dtoIn]
	if err := dtoOut.Run(dtoIn); err != nil {
		return err.Error()
	}
	return dtoOut.String()
}

// init is a method that initializes the usecases and returns a slice of interfaces
func (u *CommandUsecase) init() []interface{} {
	ret := []interface{}{}
	for k, v := range dtos {
		ret = append(ret, k)
		v.SetRepo(u.Repo)
		v.SetLog(u.Log)
	}
	return ret
}

// error is a function that logs an error and returns it
func (u *CommandUsecase) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
