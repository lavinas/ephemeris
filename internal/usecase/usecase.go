package usecase

import (
	"github.com/lavinas/ephemeris/internal/port"
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