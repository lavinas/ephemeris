package service

import (
	"fmt"
	"billing/internal/dto"
	"billing/internal/port"
)

// CreateCustomerService is responsible for handling the business logic of creating customers.
type CreateCustomerService struct {
	Base
}

// NewCreateCustomerService creates a new instance of CreateCustomerService.
func NewCreateCustomerService(repo port.Repository, logger port.Logger) *CreateCustomerService {
	return &CreateCustomerService{
		Base: *NewBase(repo, logger),
	}
}

// Run processes a batch of customer creation requests and returns the responses.
func (s *CreateCustomerService) Run(in dto.CreateCustomerRequest) dto.CreateCustomerResponse {
	s.logger.IPrintf(1, "Processing %d customer creation requests", len(in.Items))
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewCreateCustomerResponse(400, "bad request", 
		fmt.Sprintf("Validation failed: %v", err))
	}
	// get domain entities
	if err := s.repo.Save(in.GetDomain()); err != nil {
		s.logger.IPrintf(2, "Failed to save customers: %v", err)
		return dto.NewCreateCustomerResponse(500, "internal error", "contact support please")
	}
	// finalize response
	s.logger.IPrintf(1, "Successfully processed %d customer creation requests", len(in.Items))
	return dto.NewCreateCustomerResponse(200, "success", "Customers created successfully")
}

