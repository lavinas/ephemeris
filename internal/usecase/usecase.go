package usecase

import (
	"strings"

	"github.com/lavinas/ephemeris/internal/domain"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	ErrCommandNotFound = "command not identified. Please, see the help command"
)

var (
	dtos = map[interface{}]func(*Usecase, interface{}) (interface{}, string, error){
		&dto.ClientAdd{}: (*Usecase).Add,
		&dto.ClientGet{}: (*Usecase).ClientGet,
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

	// dto, err := cmd.UnmarshalOne(line, []interface{}{&dto.ClientAdd{}})
	dto, err := cmd.UnmarshalOne(line, inter)
	if err != nil {
		return err.Error()
	}
	_, str, _ := dtos[dto](u, dto)
	return str
}
