package service

import (
	"fmt"

	"billing/internal/domain"
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
func (s *CreateCustomerService) Run(in []dto.CreateCustomerRequest) dto.CreateCustomerResponse {
	s.logger.IPrintf(1, "Processing %d customer creation requests", len(in))
	// Validate input
	if len(in) == 0 {
		s.logger.IPrintf(2, "No customer data provided")
		return dto.NewCreateCustomerResponse(400, "error", "No customer data provided")
	}
	// Start transaction
	if err := s.repo.BeginTransaction(); err != nil {
		s.logger.IPrintf(2, "Failed to start transaction: %v", err)
		return dto.NewCreateCustomerResponse(500, "error", fmt.Sprintf(
			"Failed to start transaction: %v", err))
	}
	// Process each request
	for _, request := range in {
		if err := s.runOne(request); err != nil {
			s.repo.RollbackTransaction()
			return dto.NewCreateCustomerResponse(500, "Internal Error", "Contact support")
		}
	}
	// Commit transaction
	if err := s.repo.CommitTransaction(); err != nil {
		s.logger.IPrintf(2, "Failed to commit transaction: %v", err)
		return dto.NewCreateCustomerResponse(500, "error", fmt.Sprintf(
			"Failed to commit transaction: %v", err))
	}
	// finalize response
	s.logger.IPrintf(1, "Successfully processed %d customer creation requests", len(in))
	return dto.NewCreateCustomerResponse(200, "success", "Customers created successfully")
}

// RunOne processes a single customer creation request and returns the response.
func (s *CreateCustomerService) runOne(request dto.CreateCustomerRequest) error {
	s.logger.IPrintf(1, "Creating customer: %s", request.Name)
	customer := domain.NewCustomer(request.Name, request.Nickname,
		request.Document, request.Email, request.Whatsapp)
	// validate input
	if err := request.Validate(); err != nil {
		s.logger.IPrintf(1, "Erros validation %v", err)
		return err
	}
	// save customer
	if err := s.repo.Save(customer); err != nil {
		s.logger.IPrintf(1, "Error creating customer: %v", err)
		return err
	}
	// return ok
	s.logger.IPrintf(1, "Customer created successfully")
	return nil
}
