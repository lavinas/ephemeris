package service

import (
	"billing/internal/dto"
	"billing/internal/port"
	"fmt"
)

// CustomerCreate is responsible for handling the business logic of creating customers.
type CustomerCreate struct {
	Base
}

// NewCustomerCreate creates a new instance of CustomerCreate.
func NewCustomerCreate(repo port.Repository, logger port.Logger) *CustomerCreate {
	return &CustomerCreate{
		Base: *NewBase(repo, logger),
	}
}

// Run processes a batch of customer creation requests and returns the responses.
func (s *CustomerCreate) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create customer request: %v", inDTO)
	in, ok := inDTO.(*dto.CustomerCreateRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected CustomerCreateRequest")
		return dto.NewCustomerCreateResponse(400, "bad request", "Invalid input type")
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewCustomerCreateResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err))
	}
	// get domain entities
	if err := s.repo.Save(in.GetDomain()); err != nil {
		s.logger.IPrintf(2, "Failed to save customers: %v", err)
		return dto.NewCustomerCreateResponse(500, "internal error", "contact support please")
	}
	// finalize response
	s.logger.IPrintf(2, "Successfully processed %d customer creation requests", len(in.Items))
	return dto.NewCustomerCreateResponse(200, "success", "Customers created successfully")
}
