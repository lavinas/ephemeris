package service

import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
)

// InvoiceCancel handles cancelling an invoice.
type InvoiceCancel struct {
	*Base
}

// NewInvoiceCancel creates a new instance of InvoiceCancel.
func NewInvoiceCancel(repo port.Repository, logger port.Logger) *InvoiceCancel {
	return &InvoiceCancel{
		Base: NewBase(repo, logger),
	}
}

// Run executes the cancellation of an invoice.
func (s *InvoiceCancel) Run(inDTO port.InDTO) port.OutDTO {
	request, ok := inDTO.(*dto.InvoiceCancelRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type")
		return dto.NewInvoiceCancelResponse(400, "error", "Invalid input type", nil)
	}
	s.logger.IPrintf(2, "Processing invoice cancellation request: %v", request)
	if err := request.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceCancelResponse(400, "error", fmt.Sprintf("Validation failed: %v", err), nil)
	}
	domainInvoice := request.GetDomain()
	if domainInvoice == nil {
		s.logger.IPrintf(2, "Failed to convert request to domain model")
		return dto.NewInvoiceCancelResponse(500, "error", "contact support", nil)
	}
	err := s.repo.Save(domainInvoice)
	if err != nil {
		s.logger.IPrintf(2, "Failed to save cancelled invoice: %v", err)
		return dto.NewInvoiceCancelResponse(500, "error", "contact support", nil)
	}
	s.logger.IPrintf(2, "Successfully cancelled invoice with ID: %v", domainInvoice.ID)
	return dto.NewInvoiceCancelResponse(200, "success", "Invoice cancelled successfully", &domainInvoice.ID)
}
