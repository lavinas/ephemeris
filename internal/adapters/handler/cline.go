package handler

import (
	"fmt"
	"os"

	"github.com/lavinas/ephemeris/internal/port"
)

const (
	ErrNoCommandFound = "no command found"
)

type CommandLineHandler struct {
	Usecase port.CommandUseCase
}

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler(usecase port.CommandUseCase) *CommandLineHandler {
	return &CommandLineHandler{
		Usecase: usecase,
	}
}

// Run is a method that runs the command handler
func (h *CommandLineHandler) Run() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(ErrNoCommandFound)
		return
	}
	command := ""
	for _, arg := range args {
		command += arg + " "
	}
	fmt.Println(h.Usecase.Run(command[:len(command)-1]))
}
