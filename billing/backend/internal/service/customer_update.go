package service

import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
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
	request, ok := inDTO.(*dto.CustomerUpdateRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type")
		return dto.NewCustomerUpdateResponse(400, "error", "Invalid input type")
	}
	s.logger.IPrintf(2, "Processing update customer request: %v", request)
	// Validate input
	if err := request.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewCustomerUpdateResponse(400, "error", fmt.Sprintf("Validation failed: %v", err))
	}
	domainCustomer := request.GetDomain()
	if domainCustomer == nil {
		s.logger.IPrintf(2, "Failed to convert request to domain model")
		return dto.NewCustomerUpdateResponse(500, "error", "contact support")
	}
	err := s.repo.Save(domainCustomer)
	if err != nil {
		s.logger.IPrintf(2, "Failed to save customer: %v", err)
		return dto.NewCustomerUpdateResponse(500, "error", "contact support")
	}
	// Finalize response
	s.logger.IPrintf(2, "Successfully updated customer with ID: %v", domainCustomer)
	return dto.NewCustomerUpdateResponse(200, "success", "Customer updated successfully")
}
