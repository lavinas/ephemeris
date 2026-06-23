package service

import (
	"billing/internal/port"
)

// EmissionSend is responsible for handling the business logic of sending emissions.
type EmissionSend struct {
	Base
}

// NewEmissionSend creates a new instance of EmissionSend.
func NewEmissionSend(repo port.Repository, logger port.Logger, ) *EmissionSend {
	return &EmissionSend{
		Base: *NewBase(repo, logger),
	}
}

// Run processes the request to send emissions and returns the response.
func (s *EmissionSend) Run(inDTO port.InDTO) port.OutDTO {
	return nil
}
