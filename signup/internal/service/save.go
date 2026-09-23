package service

import (
	"fmt"
	"signup/internal/port"
)

// Save is responsible for handling the business logic of creating customers.
type Save struct {
	Base
}

// NewSave creates a new instance of Save.
func NewSave(repo port.Repository, logger port.Logger) *Save {
	return &Save{
		Base: *NewBase(repo, logger),
	}
}

// Run processes a batch of customer creation requests and returns the responses.
func (s *Save) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create request: %v", inDTO)
	// Type assertion to the expected DTO type
	// Validate input
	if err := inDTO.Validate(); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return inDTO.GetOutDTO(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err), nil)
	}
	domain, err := inDTO.GetDomain()
	if err != nil {
		s.logger.IPrintf(2, "Failed to get domain: %v", err)
		return inDTO.GetOutDTO(400, "bad request",
			fmt.Sprintf("Failed to get domain: %v", err), nil)
	}
	if err := domain.Validate(); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return inDTO.GetOutDTO(500, "internal error", "contact support please", nil)
	}
	if err := domain.Save(); err != nil {
		s.logger.IPrintf(2, "Failed to save customer: %v", err)
		return inDTO.GetOutDTO(500, "internal error", "contact support please", nil)
	}
	// finalize response
	s.logger.IPrintf(2, "Successfully processed customer creation requests")
	return inDTO.GetOutDTO(200, "success", "Customers created successfully", nil)
}
