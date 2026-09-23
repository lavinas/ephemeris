package service

import (
	"fmt"
	"signup/internal/port"
)

// List is responsible for handling the business logic of listing customers.
type List struct {
	Base
}

// NewList creates a new instance of List.
func NewList(repo port.Repository, logger port.Logger) *List {
	return &List{
		Base: *NewBase(repo, logger),
	}
}

// Run processes the request to list customers and returns the response.
func (s *List) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing list customer request: %v", inDTO)
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
	page, pageSize := inDTO.GetPageParams()
	found, err := domain.Find(page, pageSize)
	if err != nil {
		s.logger.IPrintf(2, "Failed to find customers: %v", err)
		return inDTO.GetOutDTO(500, "internal error", "contact support please", nil)
	}
	out := inDTO.GetOutDTO(200, "success", "Customers listed successfully", found)
	s.logger.IPrintf(2, "returned %v", out)
	return out
}