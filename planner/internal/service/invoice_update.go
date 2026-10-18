package service

import (
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceUpdate handles updating an invoice's due date and its items' invoiced quantities.
type InvoiceUpdate struct {
	Base
}

// InvoicingUpdate is an alias for InvoiceUpdate to support both nomenclatures.
type InvoicingUpdate = InvoiceUpdate

// NewInvoiceUpdate creates a new instance of InvoiceUpdate service.
func NewInvoiceUpdate(repo port.Repository, logger port.Logger) *InvoiceUpdate {
	return &InvoiceUpdate{
		Base: *NewBase(repo, logger),
	}
}

// NewInvoicingUpdate creates a new instance of InvoiceUpdate service using Invoicing nomenclature.
func NewInvoicingUpdate(repo port.Repository, logger port.Logger) *InvoiceUpdate {
	return NewInvoiceUpdate(repo, logger)
}

// Run executes the invoice update process using the provided input DTO.
func (s *InvoiceUpdate) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing update invoice request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceUpdateResponse(400, "error", err.Error(), nil)
	}

	req := input.(*dto.InvoiceUpdateRequest)

	invoice, err := s.repo.GetInvoiceByID(req.InvoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Failed to get invoice #%d: %v", req.InvoiceID, err)
		return dto.NewInvoiceUpdateResponse(500, "error", fmt.Sprintf("erro ao buscar fatura: %v", err), nil)
	}
	if invoice == nil || invoice.DeletedAt != nil {
		return dto.NewInvoiceUpdateResponse(404, "error", "pré-fatura não encontrada", nil)
	}
	if invoice.ReferenceInvoiceID != nil {
		s.logger.IPrintf(2, "Invoice #%d already invoiced (billing #%d), cannot update", invoice.ID, *invoice.ReferenceInvoiceID)
		return dto.NewInvoiceUpdateResponse(400, "error", "pré-fatura já faturada não pode ser editada", nil)
	}

	// Update DueDate
	parsedDueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return dto.NewInvoiceUpdateResponse(400, "error", "formato de data de vencimento inválido", nil)
	}
	invoice.DueDate = parsedDueDate

	// Update items invoiced quantities
	if len(req.Items) > 0 {
		itemQtyMap := make(map[int64]int)
		for _, itReq := range req.Items {
			itemQtyMap[itReq.ID] = itReq.InvoicedQuantity
		}

		for i := range invoice.Items {
			if newQty, exists := itemQtyMap[invoice.Items[i].ID]; exists {
				invoice.Items[i].InvoicedQuantity = newQty
			}
		}
	}

	// Recalculate total amount strictly based on items (Price * InvoicedQuantity)
	totalAmount := 0.0
	for _, it := range invoice.Items {
		if it.DeletedAt == nil {
			totalAmount += it.Price * float64(it.InvoicedQuantity)
		}
	}
	invoice.Amount = totalAmount

	if err := s.repo.UpdateInvoice(invoice); err != nil {
		s.logger.IPrintf(2, "Failed to update invoice in database: %v", err)
		return dto.NewInvoiceUpdateResponse(500, "error", fmt.Sprintf("erro ao salvar alterações da fatura: %v", err), nil)
	}

	s.logger.IPrintf(2, "Invoice #%d updated successfully. New amount: %.2f", invoice.ID, invoice.Amount)

	invoiceDTO := s.mapInvoiceToDTO(invoice)
	return dto.NewInvoiceUpdateResponse(200, "success", "Pré-fatura atualizada com sucesso", invoiceDTO)
}

func (s *InvoiceUpdate) mapInvoiceToDTO(inv *domain.Invoice) *dto.InvoiceDTO {
	if inv == nil {
		return nil
	}

	nickname := ""
	planType := 0
	if inv.Plan != nil {
		planType = inv.Plan.PlanType
		if inv.Plan.Customer != nil {
			nickname = inv.Plan.Customer.Nickname
		} else if inv.Plan.CustomerID != 0 && s.repo != nil {
			c, err := s.repo.GetCustomerByID(inv.Plan.CustomerID)
			if err == nil && c != nil {
				nickname = c.Nickname
			}
		}
	} else if inv.PlanID != 0 && s.repo != nil {
		p, err := s.repo.GetPlanByID(inv.PlanID)
		if err == nil && p != nil {
			planType = p.PlanType
			if p.Customer != nil {
				nickname = p.Customer.Nickname
			} else if p.CustomerID != 0 {
				c, err := s.repo.GetCustomerByID(p.CustomerID)
				if err == nil && c != nil {
					nickname = c.Nickname
				}
			}
		}
	}

	var planTypeName string
	if planType > 0 {
		planTypeName = domain.PlanType(planType).String()
	}

	items := make([]dto.InvoiceItemDTO, len(inv.Items))
	for j, it := range inv.Items {
		desc := ""
		if it.Description != nil && *it.Description != "" {
			desc = *it.Description
		} else if it.Service != nil && it.Service.Name != "" {
			desc = it.Service.Name
		} else if it.ServiceID != 0 && s.repo != nil {
			svc, err := s.repo.GetServiceByID(it.ServiceID)
			if err == nil && svc != nil {
				desc = svc.Name
			}
		}

		items[j] = dto.InvoiceItemDTO{
			ID:               it.ID,
			ServiceID:        it.ServiceID,
			Description:      desc,
			Price:            it.Price,
			InvoicedQuantity: it.InvoicedQuantity,
			RealizedQuantity: it.RealizedQuantity,
		}
	}

	periodStartStr := ""
	if !inv.PeriodStart.IsZero() {
		periodStartStr = inv.PeriodStart.Format("2006-01-02")
	}
	periodEndStr := ""
	if !inv.PeriodEnd.IsZero() {
		periodEndStr = inv.PeriodEnd.Format("2006-01-02")
	}

	return &dto.InvoiceDTO{
		ID:           inv.ID,
		InvoiceDate:  inv.InvoicingDate.Format("2006-01-02"),
		Nickname:     nickname,
		PlanType:     planType,
		PlanTypeName: planTypeName,
		DueDate:      inv.DueDate.Format("2006-01-02"),
		PeriodStart:  periodStartStr,
		PeriodEnd:    periodEndStr,
		Amount:       inv.Amount,
		Items:        items,
	}
}
