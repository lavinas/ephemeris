package service

import (
	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// CreateInvoiceService is responsible for handling the business logic of creating invoices.
type CreateInvoiceService struct {
	repo   port.Repository
	logger port.Logger
}

// NewCreateInvoiceService creates a new instance of CreateInvoiceService.
func NewCreateInvoiceService(repo port.Repository, logger port.Logger) *CreateInvoiceService {
	return &CreateInvoiceService{
		repo:   repo,
		logger: logger,
	}
}

// CreateInvoice creates a new invoice using the provided details and saves it to the repository.
func (s *CreateInvoiceService) CreateInvoice(request dto.CreateInvoiceRequest) dto.CreateInvoiceResponse {
	s.logger.IPrintf(1, "Creating invoice for customer: %s", request.CustomerName)
	invoice := domain.NewInvoice(request.CustomerName, request.CustomerEmail, request.CustomerWhatsapp, request.CustomerDocument, request.Notes,
		createItems(request.InvoiceItems))
	err := s.repo.Save(invoice)
	if err != nil {
		s.logger.IPrintf(1, "Error creating invoice: %v", err)
		return dto.CreateInvoiceResponse{
			ResponseBase: dto.ResponseBase{
				HttpCode: 500,
				Status:   "error",
				Message:  "Failed to create invoice",
			},
		}
	}
	s.logger.IPrintf(1, "Invoice created successfully")
	return dto.CreateInvoiceResponse{
		ResponseBase: dto.ResponseBase{
			HttpCode: 200,
			Status:   "success",
			Message:  "Invoice created successfully",
		},
		ID: invoice.ID,
	}
}

// createItems converts a slice of InvoiceItem DTOs to a slice of domain InvoiceItems.
func createItems(items []dto.InvoiceItem) []domain.InvoiceItem {
	invoiceItems := make([]domain.InvoiceItem, 0, len(items))
	for _, item := range items {
		invoiceItems = append(invoiceItems, domain.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}
	return invoiceItems
}
