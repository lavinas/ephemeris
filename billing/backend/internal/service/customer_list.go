package service

import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
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
	in, ok := inDTO.(*dto.CustomerListRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected CustomerListRequest")
		return dto.NewCustomerListResponse(400, "bad request", "Invalid input type", nil)
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewCustomerListResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err), nil)
	}
	// get domain entities
	customers, err := s.repo.FindCustomers(in.Page, in.PageSize, in.VendorID,
		in.Name, in.Nickname, in.Document, in.Status, in.Email, in.Whatsapp)
	if err != nil {
		s.logger.IPrintf(2, "Failed to list customers: %v", err)
		return dto.NewCustomerListResponse(500, "internal error", "contact support please", nil)
	}
	// finalize response
	customerDTOs := make([]dto.CustomerDTO, len(customers))
	for i, c := range customers {
		customerDTOs[i] = dto.NewCustomerDTO(c.ID, c.Name, c.Nickname, c.Status,
			c.Document, c.Email, c.Whatsapp)
	}
	s.logger.IPrintf(2, "Successfully listed %d customers", len(customers))
	return dto.NewCustomerListResponse(200, "success", "Customers listed successfully",
		customerDTOs)
}
