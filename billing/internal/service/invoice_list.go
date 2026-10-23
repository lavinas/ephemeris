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
		return dto.NewInvoiceListResponse(400, "bad request", "Invalid input type", nil, 0, 0)
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceListResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err), nil, 0, 0)
	}

	sentDate := in.SentDate
	if sentDate == nil && in.EmailSentDate != nil {
		sentDate = in.EmailSentDate
	}
	receiptDate := in.ReceiptDate
	if receiptDate == nil && in.EmailReceiptDate != nil {
		receiptDate = in.EmailReceiptDate
	}

	// Fetch unpaginated invoices to calculate totals
	allInvoices, err := s.repo.FindInvoices(0, 0, in.CustomerID, in.InvoiceDate,
		in.DueDate, in.PaymentDate, sentDate, receiptDate,
		in.TaxDate, in.CancellationDate, in.ExternalCode)
	if err != nil {
		s.logger.IPrintf(2, "Failed to fetch total invoices: %v", err)
		return dto.NewInvoiceListResponse(500, "internal error", "contact support please", nil, 0, 0)
	}

	// If overdue filter is specified, filter allInvoices in memory
	if in.Overdue != nil {
		filtered := make([]domain.Invoice, 0)
		for _, inv := range allInvoices {
			if inv.IsOverdue() == *in.Overdue {
				filtered = append(filtered, inv)
			}
		}
		allInvoices = filtered
	}

	totalCount := len(allInvoices)
	var totalAmount float64
	for _, inv := range allInvoices {
		totalAmount += inv.Amount
	}

	// Paginate the invoices slice
	start := (in.Page - 1) * in.PageSize
	end := start + in.PageSize
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	var invoices []domain.Invoice
	if start < totalCount {
		invoices = allInvoices[start:end]
	}

	s.logger.IPrintf(2, "Successfully fetched %d invoices (total count: %d, total amount: %.2f)", len(invoices), totalCount, totalAmount)
	return dto.NewInvoiceListResponse(200, "success", "Invoices fetched successfully",
		s.mountInvoices(invoices), totalCount, totalAmount)
}

// mountInvoices maps a slice of domain.Invoice to a slice of dto.InvoiceList for the response.
func (s *InvoiceList) mountInvoices(invoices []domain.Invoice) []dto.InvoiceList {
	responseInvoices := make([]dto.InvoiceList, len(invoices))
	for i, invoice := range invoices {
		items := make([]dto.InvoiceListListItem, len(invoice.InvoiceItems))
		for j, item := range invoice.InvoiceItems {
			items[j] = dto.NewInvoiceListListItem(item.ID, item.Description, item.Quantity, item.Price)
		}
		invDTO := dto.NewInvoiceList(invoice.ID, invoice.Customer.Nickname, invoice.Amount,
			invoice.InvoiceDate, invoice.DueDate, invoice.IsOverdue(), invoice.PaymentDate, invoice.SentDate,
			invoice.ReceiptDate, invoice.TaxDate, invoice.CancellationDate,
			invoice.Notes, invoice.ExternalCode, items)
		invDTO.CustomerName = invoice.Customer.Name
		responseInvoices[i] = invDTO
	}
	return responseInvoices
}
