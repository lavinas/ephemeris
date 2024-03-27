package handler

import (
	"fmt"
	"os"

	"github.com/lavinas/ephemeris/internal/port"
)

type CommandLineHandler struct {
	Usecase port.UseCase
}

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler(usecase port.UseCase) *CommandLineHandler {
	return &CommandLineHandler{
		Usecase: usecase,
	}
}

// Run is a method that runs the command handler
func (h *CommandLineHandler) Run() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("no command found")
		return
	}
	command := ""
	for _, arg := range args {
		command += arg + " "
	}
	fmt.Println(h.Usecase.Command(command[:len(command)-1]))
}
