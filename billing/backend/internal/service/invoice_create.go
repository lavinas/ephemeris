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
	// get domain entities
	domainInvoices, badReqErr, intErr := s.getDomain(in.Items)
	if badReqErr != nil {
		s.logger.IPrintf(2, "Bad request error: %v", badReqErr)
		return dto.NewInvoiceCreateResponse(400, "bad request", badReqErr.Error())
	}
	if intErr != nil {
		s.logger.IPrintf(2, "Internal error: %v", intErr)
		return dto.NewInvoiceCreateResponse(500, "internal error", "contact support please")
	}
	// save domain entities
	if err := s.repo.Save(domainInvoices); err != nil {
		s.logger.IPrintf(2, "Failed to save invoices: %v", err)
		return dto.NewInvoiceCreateResponse(500, "internal error", "contact support please")
	}
	s.logger.IPrintf(2, "Input validation successful for request: %v", in)
	return dto.NewInvoiceCreateResponse(200, "success", "Invoice created successfully")
}

// getDomain retrieves all invoices from the repository and returns them as a slice
// returns a slice of domain.Invoice, an error for bad request and an error for internal error
func (s *InvoiceCreate) getDomain(in []dto.InvoiceCreate) ([]domain.Invoice, error, error) {
	invoices := make([]domain.Invoice, len(in))
	for i, item := range in {
		// find vendor id by nickname
		vendor, err := s.repo.GetVendor(item.Vendor)
		if err != nil {
			return nil, nil, fmt.Errorf("error finding vendor: %v", err)
		}
		if vendor == nil {
			return nil, fmt.Errorf("vendor not found: %s", item.Vendor), nil
		}
		businessID := vendor.ID
		// find customer id
		customer, err := s.repo.GetCustomer(item.Customer)
		if err != nil {
			return nil, nil, fmt.Errorf("error finding customer: %v", err)
		}
		if customer == nil {
			return nil, fmt.Errorf("customer not found: %s", item.Customer), nil
		}
		customerID := customer.ID
		invoices[i] = *item.GetDomain(businessID, customerID)
	}
	return invoices, nil, nil
}
