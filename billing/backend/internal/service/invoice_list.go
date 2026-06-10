package service

import (
	"fmt"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// InvoiceList is responsible for handling the business logic of listing invoices.
type InvoiceList struct {
	Base
}

// NewInvoiceList creates a new instance of InvoiceList.
func NewInvoiceList(repo port.Repository, logger port.Logger) *InvoiceList {
	return &InvoiceList{
		Base: *NewBase(repo, logger),
	}
}

// Run processes the request to list invoices and returns the response.
func (s *InvoiceList) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing list invoices request: %v", inDTO)
	// Type assertion to the expected DTO type
	in, ok := inDTO.(*dto.InvoiceListRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected InvoiceListRequest")
		return dto.NewInvoiceListResponse(400, "bad request", "Invalid input type", nil)
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceListResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err), nil)
	}
	// Fetch invoices from the repository
	invoices, err := s.repo.FindInvoices(in.Page, in.PageSize, in.CustomerID, in.InvoiceDate,
		in.DueDate, in.EmailSentDate, in.WhatsappSentDate, in.TaxDate)
	if err != nil {
		s.logger.IPrintf(2, "Failed to fetch invoices: %v", err)
		return dto.NewInvoiceListResponse(500, "internal error", "contact support please", nil)
	}
	s.logger.IPrintf(2, "Successfully fetched %d invoices", len(invoices))
	return dto.NewInvoiceListResponse(200, "success", "Invoices fetched successfully",
		s.mountInvoices(invoices))
}

// mountInvoices maps a slice of domain.Invoice to a slice of dto.InvoiceList for the response.
func (s *InvoiceList) mountInvoices(invoices []domain.Invoice) []dto.InvoiceList {
	responseInvoices := make([]dto.InvoiceList, len(invoices))
	for i, invoice := range invoices {
		items := make([]dto.InvoiceListListItem, len(invoice.InvoiceItems))
		for j, item := range invoice.InvoiceItems {
			items[j] = dto.NewInvoiceListListItem(item.ID, item.Description, item.Quantity, item.Price)
		}
		responseInvoices[i] = dto.NewInvoiceList(invoice.ID, invoice.Customer.Nickname, invoice.Amount,
			invoice.InvoiceDate, invoice.DueDate, invoice.EmailSentDate, invoice.WhatsappSentDate,
			invoice.TaxDate, invoice.Notes, items)
	}
	return responseInvoices
}
