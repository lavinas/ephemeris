package usecase

import (
	"errors"
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

// Usecase is a struct that groups all usecases of the application
type CommandUsecase struct {
	Repo    port.Repository
	Log     port.Logger
	UseCase port.UseCase
}

// UseCase is a function that returns a new UseCase struct
func NewCommandUsecase(repo port.Repository, log port.Logger) *CommandUsecase {
	if err := repo.Migrate(domain.All()); err != nil {
		panic(err)
	}
	return &CommandUsecase{
		Repo:    repo,
		Log:     log,
		UseCase: NewUsecase(repo, log),
	}
}

// Run is a method that runs a command
func (u *CommandUsecase) Run(line string) string {
	u.Log.Println("Command: " + line)
	line = strings.ToLower(line)
	cmd := pkg.Commands{}
	dtoIn, err := cmd.FindOne(line, dto.All())
	if err != nil {
		return u.error(pkg.ErrPrefCommandNotFound, err.Error()).Error()
	}
	if err := cmd.Unmarshal(line, dtoIn); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error()).Error()
	}
	if err := u.UseCase.Run(dtoIn); err != nil {
		return err.Error()
	}
	return u.UseCase.String()
}

// error is a function that logs an error and returns it
func (u *CommandUsecase) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
