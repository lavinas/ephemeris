package service

import (

	"billing/internal/port"
	"billing/internal/dto"
)

// EmissionReceive is responsible for handling the business logic of receiving emissions.
type EmissionReceive struct {
	Base
	receiver port.Issuer
}

// NewEmissionReceive creates a new instance of EmissionReceive.
func NewEmissionReceive(repo port.Repository, logger port.Logger, issuer port.Issuer) *EmissionReceive {
	return &EmissionReceive{
		Base:     *NewBase(repo, logger),
		receiver: issuer,
	}
}

// Run processes the request to receive emissions and returns the response.
func (s *EmissionReceive) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(1, "Running EmissionReceive service")
	receiveDTO, ok := inDTO.(*dto.EmissionReceiveRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type")
		return dto.NewEmissionReceiveResponse(400, "error", "Invalid input type", 0, 0, 0)
	}
	if err := receiveDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation error: %v", err)
		return dto.NewEmissionReceiveResponse(400, "error", err.Error(), 0, 0, 0)
	}
	// Process the emission receive logic here
	s.logger.IPrintf(1, "Emission received successfully: ID %d, Quantity %d, Amount %.2f",
		receiveDTO.Emission.ID, receiveDTO.Emission.Quantity, receiveDTO.Emission.Amount)
	// Return a successful response
	return dto.NewEmissionReceiveResponse(200, "success", "Emission received successfully",
		receiveDTO.Emission.ID, receiveDTO.Emission.Quantity, receiveDTO.Emission.Amount)
}