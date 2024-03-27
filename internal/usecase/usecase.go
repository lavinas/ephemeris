package usecase

import (
	"reflect"
	"strings"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	ErrCommandNotFound = "command not identified. Please, see the help command"
	Err
)

var (
	dtos = map[port.DTO]func(*Usecase, port.DTO) (string, error){
		reflect.TypeOf(dto.ClientAdd{}): (*Usecase).ClientAdd,
		reflect.TypeOf(dto.ClientGet{}): (*Usecase).ClientGet,
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
	dto, err := cmd.UnmarshalOne(line, []interface{}{&dto.ClientAdd{}, &dto.ClientGet{}})
	if err != nil {
		return err.Error()
	}
	str, _ := dtos[reflect.TypeOf(dto).Elem()](u, dto)
	return str
}
