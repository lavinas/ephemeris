package service

import (
	"fmt"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// InvoiceCreate is responsible for handling the business logic of creating a new invoice.
type InvoiceCreate struct {
	Base
}

// NewInvoiceCreate creates a new instance of InvoiceCreate.
func NewInvoiceCreate(repo port.Repository, logger port.Logger) *InvoiceCreate {
	return &InvoiceCreate{
		Base: *NewBase(repo, logger),
	}
}

// Run processes the request to create a new invoice and returns the response.
func (s *InvoiceCreate) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create invoice request: %v", inDTO)
	in, ok := inDTO.(*dto.InvoiceCreateRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected InvoiceCreateRequest")
		return dto.NewInvoiceCreateResponse(400, "bad request", "Invalid input type")
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceCreateResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err))
	}
	return dto.NewInvoiceCreateResponse(200, "success", "Invoice created successfully")
}

// getDomain retrieves all invoices from the repository and returns them as a slice
// returns a slice of domain.Invoice, an error for bad request and an error for internal error
func (s *InvoiceCreate) getDomain(in []dto.InvoiceCreate) ([]domain.Invoice, error, error) {
	invoices := make([]domain.Invoice, len(in))
	for i, item := range in {
		// find vendor id by nickn
		b, err := s.repo.FindVendors(1, 1, nil, &item.Business, nil, nil, nil, nil)
		if err != nil {
			s.logger.IPrintf(2, "Error finding vendor: %v", err)
			return nil, fmt.Errorf("error finding vendor: %v", err), nil
		}
		if len(b) == 0 {
			s.logger.IPrintf(2, "Vendor not found: %s", item.Business)
			return nil, fmt.Errorf("vendor not found: %s", item.Business), nil
		}
		businessID := b[0].ID
		// find customer id
		c, err := s.repo.FindCustomers(1, 1, &item.Customer, nil, nil, nil, nil, nil)
		if err != nil {
			s.logger.IPrintf(2, "Error finding customer: %v", err)
			return nil, fmt.Errorf("error finding customer: %v", err), nil
		}
		if len(c) == 0 {
			s.logger.IPrintf(2, "Customer not found: %s", item.Customer)
			return nil, fmt.Errorf("customer not found: %s", item.Customer), nil
		}
		customerID := c[0].ID
		invoices[i] = *item.GetDomain(businessID, customerID)
	}
	return invoices, nil, nil
}
