package usecase

import (
	"fmt"
	"strings"

	"github.com/lavinas/ephemeris/internal/port"
)

const (
	ErrorCommandShort = "command should have at least 1 parameter"
	ErrorCommandNotFound = "command not found: %s, possible commands: %s"
)

var (
	// cmds is a map that groups all starting commands
	cmds = map[string]func(*Usecase, string) string{
		"client":  (*Usecase).CommandClient,
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

// Command is a method that receives a command and execute it
func (c *Usecase) Command(cmd string) string {
	cmd = strings.ToLower(cmd)
	c.Log.Println(cmd)
	cmdSlice := strings.Split(cmd, " ")
	if len(cmdSlice) == 0 {
		return ErrorCommandShort
	}
	if f, ok := cmds[cmdSlice[0]]; ok {
		cmd := strings.Join(cmdSlice[1:], " ")
		return f(c, cmd)
	}		
	return fmt.Sprintf(ErrorCommandNotFound, cmdSlice[0], cmdString(cmds))
}	

// getCommands is a method that returns all possible commands
func cmdString[K string, V any](map[K]V) string {
	commands := make([]string, 0, len(cmds))
	for k := range cmds {
		commands = append(commands, k)
	}
	return strings.Join(commands, ", ")
}


// MapCommand is a function that returns a map of the command
func MapCommand(cmd string)map[string]string {
	cmd = strings.ToLower(cmd)
	cmdSlice := strings.Split(cmd, " ")
	maps := make(map[string]string)
	for _, f := range cmdSlice {
		if f[0:1] == "'" || f[0:1] == "\"" {
			continue
		}
		maps[f] = ""
	}
	return maps
}