package service

import (
	"fmt"

	"signup/internal/dto"
	"signup/internal/port"
)

type CustomerUpdate struct {
	*Base
}

// NewCustomerUpdate creates a new instance of CustomerUpdate with the provided repository and logger.
func NewCustomerUpdate(repo port.Repository, logger port.Logger) *CustomerUpdate {
	return &CustomerUpdate{
		Base: NewBase(repo, logger),
	}
}

// Run executes the customer update process using the provided request data and returns a response.
func (s *CustomerUpdate) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing update customer request: %v", inDTO)
	// Validate input
	if err := inDTO.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewCustomerUpdateResponse(400, "error", fmt.Sprintf("Validation failed: %v", err))
	}
	domain := inDTO.GetDomain()
	err := s.repo.Save(domain)
	if err != nil {
		s.logger.IPrintf(2, "Failed to save customer: %v", err)
		return dto.NewCustomerUpdateResponse(500, "error", "contact support")
	}
	// Finalize response
	s.logger.IPrintf(2, "Successfully updated customer with ID: %v", domainCustomer)
	return dto.NewCustomerUpdateResponse(200, "success", "Customer updated successfully")
}
