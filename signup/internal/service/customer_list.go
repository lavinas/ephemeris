package service

import (
	"fmt"
	"signup/internal/port"
)

// CustomerList is responsible for handling the business logic of listing customers.
type CustomerList struct {
	Base
}

// NewCustomerList creates a new instance of CustomerList.
func NewCustomerList(repo port.Repository, logger port.Logger) *CustomerList {
	return &CustomerList{
		Base: *NewBase(repo, logger),
	}
}

// Run processes the request to list customers and returns the response.
func (s *CustomerList) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing list customer request: %v", inDTO)
	// Validate input
	if err := inDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return inDTO.GetOutDTO(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err), nil)
	}
	domain := inDTO.GetDomain()
	found, err := domain.Find(s.repo)
	if err != nil {
		s.logger.IPrintf(2, "Failed to find customers: %v", err)
		return inDTO.GetOutDTO(500, "internal error", "contact support please", nil)
	}
	return inDTO.GetOutDTO(200, "success", "Customers listed successfully", found)
}
