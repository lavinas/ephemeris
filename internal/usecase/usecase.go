package usecase

import (
	"errors"
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	ErrPrefBadRequest      = "bad request"
	ErrPrefCommandNotFound = "command not identified"
	ErrPrefInternal        = "internal error"
	ErrPrefConflict        = "conflict"
	ErrCommandNotFound     = "command not identified. Please, see the help command"
)

var (
	dtos = map[interface{}]func(*Usecase, port.DTO) (interface{}, string, error){
		&dto.ClientAdd{}:  (*Usecase).Add,
		&dto.ClientAdd2{}: (*Usecase).Add,
		&dto.ClientGet{}:  (*Usecase).Get,
		&dto.ClientGet2{}: (*Usecase).Get,
		&dto.ClientUp{}:   (*Usecase).Up,
		&dto.ClientUp2{}:  (*Usecase).Up,
	}
)

// Usecase is a struct that groups all usecases of the application
type Usecase struct {
	Repo port.Repository
	Log  port.Logger
	Cfg  port.Config
}

// UseCase is a function that returns a new UseCase struct
func NewUsecase(repo port.Repository, log port.Logger) *Usecase {
	repo.Migrate(domain.GetDomain())
	return &Usecase{
		Repo: repo,
		Log:  log,
	}
}

// GetDTO is a function that converts a string command to a DTO
func (u *Usecase) Command(line string) string {
	u.Log.Println("Command: " + line)
	line = strings.ToLower(line)
	cmd := pkg.Commands{}
	inter := []interface{}{}
	for k := range dtos {
		inter = append(inter, k)
	}
	dto, err := cmd.FindOne(line, inter)
	if err != nil {
		return u.error(ErrPrefCommandNotFound, err.Error()).Error()
	}
	if err := cmd.Unmarshal(line, dto); err != nil {
		return u.error(ErrPrefBadRequest, err.Error()).Error()
	}
	dtx := dto.(port.DTO)
	_, str, _ := dtos[dto](u, dtx)
	return str
}

// error is a function that logs an error and returns it
func (u *Usecase) error(prefix string, err string) error {
	err = prefix + ": " + err
	u.Log.Println(err)
	return errors.New(err)
}
