package usecase

import (
	"github.com/lavinas/ephemeris/internal/port"
)

// ClientUsecase is a struct that defines the usecase of the client
type ClientUsecase struct {
	Repo port.Repository
	Log  port.Logger
	Cfg  port.Config
}

// Insert is a method that inserts a client
func (c *ClientUsecase) Register(id string, name string, document string, email string, phone string, contactWay string, role string, bond string) {
	c.Log.Info("Getting client")
}
