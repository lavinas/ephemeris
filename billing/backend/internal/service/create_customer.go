package service

import (
	"billing/internal/dto"
	"billing/internal/port"
	"fmt"
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
func (s *CreateCustomerService) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create customer request: %v", inDTO)
	in, ok := inDTO.(*dto.CreateCustomerRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected CreateCustomerRequest")
		return dto.NewCreateCustomerResponse(400, "bad request", "Invalid input type")
	}
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
	s.logger.IPrintf(2, "Successfully processed %d customer creation requests", len(in.Items))
	return dto.NewCreateCustomerResponse(200, "success", "Customers created successfully")
}
