package usecase

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

const (
	ErrorCommandShort     = "wrong command. Should have: object action <paramenters>. Ex: client get nickname. Use help for more information"
	ErrorObjectAction     = "wrong object or action. Use help for more information"
	ErrorMissingParameter = "missing parameter: %s"
)

var (
	dtos = map[port.DTO]func(*Usecase, port.DTO) error{
		&dto.ClientAdd{Base: dto.Base{Object: "client", Action: "add"}}: (*Usecase).AddClient,
		&dto.ClientGet{Base: dto.Base{Object: "client", Action: "get"}}: (*Usecase).GetClient,
	}
)

// Usecase is a struct that groups all usecases of the application
type Usecase struct {
	Repo port.Repository
	Log  port.Logger
	Cfg  port.Config
}

// NewClientUsecase is a function that returns a new ClientUsecase
func NewClientUsecase(repo port.Repository, log port.Logger) *Usecase {
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
	
	




	cmdSlice := strings.Split(cmd, " ")
	if len(cmdSlice) < 2 {
		return ErrorCommandShort
	}
	for dto, f := range dtos {
		if slices.Contains(cmdSlice, dto.GetObject()) && slices.Contains(cmdSlice, dto.GetAction()) {
			if err := u.getparams(dto, cmdSlice); err != nil {
				u.Log.Println(err.Error())
				return err.Error()
			}
			if err := f(u, dto); err != nil {
				u.Log.Println(err.Error())
				return err.Error()
			}
		}
	}
	u.Log.Println(ErrorObjectAction)
	return ErrorObjectAction
}

// getparams is a function that converts a string command to a DTO
func (u *Usecase) getparams(dto port.DTO, cmdSlice []string) error {
	st := reflect.TypeOf(dto).Elem()
	tags := u.commandGetTags(st)
	values, err := u.commandMapValues(tags, cmdSlice)
	if err != nil {
		return err
	}
	u.commandSetFields(dto, values)
	return nil
}

// getCommandTags is a function that returns all tags of a struct
func (u *Usecase) commandGetTags(st reflect.Type) map[string]string {
	tags := map[string]string{}
	for i := 0; i < st.NumField(); i++ {
		tag := st.Field(i).Tag.Get("command")
		if tag == "" {
			continue
		}
		tags[tag] = st.Field(i).Name
	}
	return tags
}

// mapCommandValues is a function that maps command fields values
func (u *Usecase) commandMapValues(alltags map[string]string, cmdSlice []string) (map[string]string, error) {
	valueMap := map[string]string{}
	for tag, field := range alltags {
		if !slices.Contains(cmdSlice, tag) {
			return valueMap, fmt.Errorf(ErrorMissingParameter, tag)
		}
		param := ""
		for j := slices.Index(cmdSlice, tag) + 1; j < len(cmdSlice); j++ {
			if alltags[cmdSlice[j]] != "" {
				break
			}
			param += cmdSlice[j] + " "
		}
		valueMap[field] = param
	}
	return valueMap, nil
}

// setCommandFields is a function that sets the fields of a DTO
func (u *Usecase) commandSetFields(dto port.DTO, values map[string]string) {
	for k, v := range values {
		field := reflect.ValueOf(dto).Elem().FieldByName(k)
		field.SetString(v)
	}

}
