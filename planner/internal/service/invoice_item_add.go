package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceItemAdd handles adding a standalone item (addition or discount) to an existing pre-invoice.
type InvoiceItemAdd struct {
	Base
	billingClient port.BillingClient
	comptime      port.ComptimeService
}

// InvoicingItemAdd is an alias for InvoiceItemAdd.
type InvoicingItemAdd = InvoiceItemAdd

// NewInvoiceItemAdd creates a new instance of InvoiceItemAdd service.
func NewInvoiceItemAdd(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *InvoiceItemAdd {
	var client port.BillingClient
	if len(billingClients) > 0 && billingClients[0] != nil {
		client = billingClients[0]
	}
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &InvoiceItemAdd{
		Base:          *NewBase(repo, logger),
		billingClient: client,
		comptime:      comptime,
	}
}

// NewInvoicingItemAdd creates a new instance of InvoiceItemAdd service using Invoicing nomenclature.
func NewInvoicingItemAdd(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *InvoiceItemAdd {
	return NewInvoiceItemAdd(repo, logger, billingClients...)
}

// SetComptimeService sets the comptime service.
func (s *InvoiceItemAdd) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}

// Run executes adding a standalone item to a pre-invoice.
func (s *InvoiceItemAdd) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing add standalone invoice item request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceItemAddResponse(400, "error", err.Error(), nil, nil)
	}

	req, ok := input.(*dto.InvoiceItemAddRequest)
	if !ok {
		return dto.NewInvoiceItemAddResponse(400, "error", "payload de entrada inválido", nil, nil)
	}

	req.ApplyDefaults()

	pStart, err := dto.ParseFlexibleDate(req.PeriodStart)
	if err != nil {
		return dto.NewInvoiceItemAddResponse(400, "error", fmt.Sprintf("período inicial inválido: %v", err), nil, nil)
	}
	pEnd, err := dto.ParseFlexibleDate(req.PeriodEnd)
	if err != nil {
		return dto.NewInvoiceItemAddResponse(400, "error", fmt.Sprintf("período final inválido: %v", err), nil, nil)
	}

	invoice, err := s.repo.GetInvoiceByID(req.InvoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Failed to get invoice #%d: %v", req.InvoiceID, err)
		return dto.NewInvoiceItemAddResponse(500, "error", fmt.Sprintf("erro ao buscar fatura: %v", err), nil, nil)
	}
	if invoice == nil || invoice.DeletedAt != nil {
		return dto.NewInvoiceItemAddResponse(404, "error", "pré-fatura não encontrada", nil, nil)
	}

	var svc *domain.Service
	if req.ServiceID > 0 {
		svc, _ = s.repo.GetServiceByID(req.ServiceID)
	}
	if svc == nil {
		services, err := s.repo.FindServices(1)
		if err != nil {
			return dto.NewInvoiceItemAddResponse(500, "error", fmt.Sprintf("erro ao buscar serviços: %v", err), nil, nil)
		}
		for i := range services {
			if (req.ServiceID > 0 && services[i].ID == req.ServiceID) ||
				(req.ServiceName != "" && strings.EqualFold(services[i].Name, strings.TrimSpace(req.ServiceName))) {
				svc = &services[i]
				break
			}
		}
	}
	if svc == nil {
		return dto.NewInvoiceItemAddResponse(404, "error", "serviço não encontrado", nil, nil)
	}

	descStr := strings.TrimSpace(req.Description)
	if descStr == "" {
		baseDesc := svc.Name
		if svc.Description != nil && strings.TrimSpace(*svc.Description) != "" {
			baseDesc = strings.TrimSpace(*svc.Description)
		}
		if req.Price < 0 {
			descStr = fmt.Sprintf("Desconto %s - %s a %s", baseDesc, pStart.Format("02/01/2006"), pEnd.Format("02/01/2006"))
		} else {
			descStr = fmt.Sprintf("%s - %s a %s", baseDesc, pStart.Format("02/01/2006"), pEnd.Format("02/01/2006"))
		}
	}

	newItem := domain.NewInvoiceItem(
		invoice.ID,
		svc.ID,
		req.Price,
		&descStr,
		req.Quantity,
		req.Quantity,
		pStart,
		pEnd,
	)
	newItem.Quantity = req.Quantity
	newItem.Service = svc

	invoice.Items = append(invoice.Items, *newItem)

	// Recalculate total amount across all active items
	totalAmount := 0.0
	for _, it := range invoice.Items {
		if it.DeletedAt == nil {
			qty := it.InvoicedQuantity
			if qty == 0 && it.Quantity > 0 {
				qty = it.Quantity
			}
			totalAmount += it.Price * float64(qty)
		}
	}
	invoice.Amount = totalAmount

	// If the invoice was already billed in billing, cancel previous billing invoice and create replacement
	if invoice.ReferenceInvoiceID != nil && s.billingClient != nil {
		if err := syncUpdatedInvoiceWithBilling(s.repo, s.billingClient, s.logger, invoice); err != nil {
			return dto.NewInvoiceItemAddResponse(500, "error", err.Error(), nil, nil)
		}
	}

	if err := s.repo.UpdateInvoice(invoice); err != nil {
		s.logger.IPrintf(2, "Failed to update invoice with new item in database: %v", err)
		return dto.NewInvoiceItemAddResponse(500, "error", fmt.Sprintf("erro ao salvar item na pré-fatura: %v", err), nil, nil)
	}

	createdItem := &invoice.Items[len(invoice.Items)-1]
	if s.comptime != nil && createdItem.ID > 0 {
		if err := s.comptime.ProcessInvoiceItemCreated(context.Background(), createdItem.ID); err != nil {
			s.logger.IPrintf(2, "Aviso: falha ao processar novo item avulso %d no banco de horas: %v", createdItem.ID, err)
		}
	}

	s.logger.IPrintf(2, "Standalone item #%d added to invoice #%d. New amount: %.2f", createdItem.ID, invoice.ID, invoice.Amount)

	updater := NewInvoiceUpdate(s.repo, s.logger, s.billingClient)
	invoiceDTO := updater.mapInvoiceToDTO(invoice)
	var itemDTO *dto.InvoiceItemDTO
	if invoiceDTO != nil && len(invoiceDTO.Items) > 0 {
		last := invoiceDTO.Items[len(invoiceDTO.Items)-1]
		itemDTO = &last
	}

	return dto.NewInvoiceItemAddResponse(201, "success", "Item avulso adicionado com sucesso", invoiceDTO, itemDTO)
}

func syncUpdatedInvoiceWithBilling(repo port.Repository, billingClient port.BillingClient, logger port.Logger, invoice *domain.Invoice) error {
	vendorNickname, customerNickname, err := resolveInvoiceVendorAndCustomer(repo, invoice)
	if err != nil {
		logger.IPrintf(2, "Failed to resolve vendor/customer for invoice #%d billing update: %v", invoice.ID, err)
		return fmt.Errorf("erro ao resolver cliente/vendedor para faturamento: %v", err)
	}

	oldRefID := *invoice.ReferenceInvoiceID
	todayStr := time.Now().Format("2006-01-02")
	extCode := fmt.Sprintf("%d", invoice.ID)
	cancelInput := port.BillingCancelInput{
		Vendor:           vendorNickname,
		ID:               oldRefID,
		ExternalCode:     &extCode,
		CancellationDate: &todayStr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := billingClient.CancelInvoice(ctx, cancelInput); err != nil {
		logger.IPrintf(2, "Failed to cancel previous invoice %d in billing (ref #%d): %v", invoice.ID, oldRefID, err)
		return fmt.Errorf("erro ao cancelar fatura anterior no billing: %v", err)
	}

	items := make([]port.BillingItemInput, 0, len(invoice.Items))
	for _, it := range invoice.Items {
		if it.DeletedAt != nil {
			continue
		}
		desc := ""
		if it.Description != nil && *it.Description != "" {
			desc = *it.Description
		} else if it.Service != nil && it.Service.Name != "" {
			desc = it.Service.Name
		} else if it.ServiceID != 0 && repo != nil {
			svc, err := repo.GetServiceByID(it.ServiceID)
			if err == nil && svc != nil {
				desc = svc.Name
			}
		}
		if desc == "" {
			desc = fmt.Sprintf("Serviço #%d", it.ServiceID)
		}

		qty := it.InvoicedQuantity
		if qty <= 0 && it.Quantity > 0 {
			qty = it.Quantity
		}
		if qty <= 0 {
			qty = 1
		}

		items = append(items, port.BillingItemInput{
			Description: desc,
			Price:       it.Price,
			Quantity:    qty,
		})
	}

	billingInput := port.BillingInvoiceInput{
		Vendor:       vendorNickname,
		Customer:     customerNickname,
		Invoicing:    invoice.InvoicingDate.Format("2006-01-02"),
		Due:          invoice.DueDate.Format("2006-01-02"),
		ExternalCode: &extCode,
		Amount:       invoice.Amount,
		Items:        items,
	}

	newRefID, err := billingClient.CreateInvoice(ctx, billingInput)
	if err != nil {
		logger.IPrintf(2, "Failed to create replacement invoice in billing for pre-invoice %d: %v", invoice.ID, err)
		return fmt.Errorf("fatura anterior cancelada no billing, mas erro ao criar nova fatura: %v", err)
	}
	invoice.ReferenceInvoiceID = &newRefID
	return nil
}
