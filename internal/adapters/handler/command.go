package handler

import (
	"github.com/lavinas/ephemeris/internal/port"
)

type CommandHandler struct {
	Usecase port.UseCase
}

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler(usecase port.UseCase) *CommandHandler {
	return &CommandHandler{
		Usecase: usecase,
	}
}

func (h *CommandHandler) Run() {

}
