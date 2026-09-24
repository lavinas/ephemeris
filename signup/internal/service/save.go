package service

import (
	"context"
	"fmt"
	"signup/internal/port"
)

// Save is responsible for handling the business logic of creating and updating entities.
type Save struct {
	Base
	publisher port.CustomerEventPublisher
}

// NewSave creates a new instance of Save.
func NewSave(repo port.Repository, logger port.Logger, publisher port.CustomerEventPublisher) *Save {
	return &Save{
		Base:      *NewBase(repo, logger),
		publisher: publisher,
	}
}

// Run processes an entity creation or update request and returns the response.
func (s *Save) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing save request: %v", inDTO)
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
		s.logger.IPrintf(2, "Failed to save entity: %v", err)
		return inDTO.GetOutDTO(500, "internal error", "contact support please", nil)
	}

	// If the DTO implements EventEmittable and publisher is provided, emit event
	if s.publisher != nil {
		if emittable, ok := inDTO.(port.EventEmittable); ok {
			if err := emittable.EmitEvent(context.Background(), s.publisher); err != nil {
				s.logger.IPrintf(1, "Warning: failed to publish domain event: %v", err)
			}
		}
	}

	// finalize response
	s.logger.IPrintf(2, "Successfully processed save request")
	return inDTO.GetOutDTO(200, "success", "Saved successfully", nil)
}
