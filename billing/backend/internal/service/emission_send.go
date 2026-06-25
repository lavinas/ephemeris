package service

import (
	"billing/internal/port"
)

// EmissionSend is responsible for handling the business logic of sending emissions.
type EmissionSend struct {
	Base
	sender port.EmissionSender
}

// NewEmissionSend creates a new instance of EmissionSend.
func NewEmissionSend(repo port.Repository, logger port.Logger, emissionSender port.EmissionSender) *EmissionSend {
	return &EmissionSend{
		Base:   *NewBase(repo, logger),
		sender: emissionSender,
	}
}

// Run processes the request to send emissions and returns the response.
func (s *EmissionSend) Run(inDTO port.InDTO) port.OutDTO {
	return nil
}
