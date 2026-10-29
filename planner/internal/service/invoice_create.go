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

// InvoiceCreate is a service that handles creating manual invoices.
type InvoiceCreate struct {
	Base
	comptime port.ComptimeService
}

// NewInvoiceCreate creates a new instance of InvoiceCreate service.
func NewInvoiceCreate(repo port.Repository, logger port.Logger) *InvoiceCreate {
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &InvoiceCreate{
		Base:     *NewBase(repo, logger),
		comptime: comptime,
	}
}

// SetComptimeService sets the comptime service.
func (s *InvoiceCreate) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}

// Run executes the manual invoice creation process.
func (s *InvoiceCreate) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing manual invoice create request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Invoice create validation error: %v", err)
		return dto.NewInvoiceCreateResponse(400, "Bad Request", err.Error(), nil)
	}

	req, ok := input.(*dto.InvoiceCreateRequest)
	if !ok {
		return dto.NewInvoiceCreateResponse(400, "Bad Request", "payload de entrada inválido", nil)
	}

	req.ApplyDefaults()

	invoiceDate, err := dto.ParseFlexibleDate(req.InvoiceDate)
	if err != nil {
		return dto.NewInvoiceCreateResponse(400, "Bad Request", fmt.Sprintf("data da fatura inválida: %v", err), nil)
	}
	dueDate, err := dto.ParseFlexibleDate(req.DueDate)
	if err != nil {
		return dto.NewInvoiceCreateResponse(400, "Bad Request", fmt.Sprintf("data de vencimento inválida: %v", err), nil)
	}
	periodStart, err := dto.ParseFlexibleDate(req.PeriodStart)
	if err != nil {
		return dto.NewInvoiceCreateResponse(400, "Bad Request", fmt.Sprintf("período inicial inválido: %v", err), nil)
	}
	periodEnd, err := dto.ParseFlexibleDate(req.PeriodEnd)
	if err != nil {
		return dto.NewInvoiceCreateResponse(400, "Bad Request", fmt.Sprintf("período final inválido: %v", err), nil)
	}

	customer, err := s.repo.GetCustomer(1, req.Nickname)
	if err != nil || customer == nil {
		return dto.NewInvoiceCreateResponse(404, "Not Found", fmt.Sprintf("cliente '%s' não encontrado", req.Nickname), nil)
	}

	services, err := s.repo.FindServices(1)
	if err != nil {
		return dto.NewInvoiceCreateResponse(500, "error", fmt.Sprintf("erro ao buscar serviços: %v", err), nil)
	}

	for i := range req.Items {
		if req.Items[i].ServiceID == 0 && req.Items[i].ServiceName != "" {
			for _, svc := range services {
				if strings.EqualFold(svc.Name, req.Items[i].ServiceName) {
					req.Items[i].ServiceID = svc.ID
					break
				}
			}
		}
	}

	serviceIDs := make([]int64, 0, len(req.Items))
	for _, it := range req.Items {
		serviceIDs = append(serviceIDs, it.ServiceID)
	}

	planID, err := s.resolveOrCreatePlan(customer.ID, invoiceDate)
	if err != nil {
		return dto.NewInvoiceCreateResponse(500, "error", fmt.Sprintf("erro ao associar plano: %v", err), nil)
	}

	totalAmount := 0.0
	items := make([]domain.InvoiceItem, len(req.Items))
	for i, it := range req.Items {
		totalAmount += it.Price * float64(it.Quantity)
		var desc *string
		if strings.TrimSpace(it.Description) != "" {
			d := strings.TrimSpace(it.Description)
			desc = &d
		}
		items[i] = domain.InvoiceItem{
			ServiceID:   it.ServiceID,
			Price:       it.Price,
			Description: desc,
			Quantity:    it.Quantity,
		}
	}

	invoice := domain.NewInvoice(planID, invoiceDate, dueDate, periodStart, periodEnd, totalAmount, req.ReferenceInvoiceID)
	invoice.Items = items

	if err := s.repo.SaveInvoice(invoice); err != nil {
		s.logger.IPrintf(2, "Failed to save invoice: %v", err)
		return dto.NewInvoiceCreateResponse(500, "error", fmt.Sprintf("erro ao salvar fatura: %v", err), nil)
	}

	if s.comptime != nil {
		for _, it := range invoice.Items {
			if err := s.comptime.ProcessInvoiceItemCreated(context.Background(), it.ID); err != nil {
				s.logger.IPrintf(2, "Aviso: falha ao processar item %d no banco de horas: %v", it.ID, err)
			}
		}
	}

	s.logger.IPrintf(2, "Manual invoice #%d created successfully for customer %s. Total: %.2f", invoice.ID, customer.Nickname, invoice.Amount)

	invoiceDTO := s.mapInvoiceToDTO(invoice, customer.Nickname)
	return dto.NewInvoiceCreateResponse(201, "success", "Fatura criada com sucesso", invoiceDTO)
}

func (s *InvoiceCreate) resolveOrCreatePlan(customerID int64, invoiceDate time.Time) (int64, error) {
	conditions := map[string]interface{}{
		"customer_id = ?":                     customerID,
		"deleted_at IS NULL":                  nil,
		"plan_type = ?":                       int(domain.PlanTypeManual),
		"plan_start <= ?":                     invoiceDate,
		"(plan_end IS NULL OR plan_end >= ?)": invoiceDate,
	}

	plans, _, err := s.repo.FindPlans(1, 100, conditions, "id desc")
	if err == nil {
		invDateOnly := time.Date(invoiceDate.Year(), invoiceDate.Month(), invoiceDate.Day(), 0, 0, 0, 0, time.UTC)
		for _, p := range plans {
			if p.CustomerID == customerID && p.DeletedAt == nil && p.PlanType == int(domain.PlanTypeManual) {
				pStart := time.Date(p.PlanStart.Year(), p.PlanStart.Month(), p.PlanStart.Day(), 0, 0, 0, 0, time.UTC)
				matches := !invDateOnly.Before(pStart)
				if p.PlanEnd != nil {
					pEnd := time.Date(p.PlanEnd.Year(), p.PlanEnd.Month(), p.PlanEnd.Day(), 0, 0, 0, 0, time.UTC)
					matches = matches && !invDateOnly.After(pEnd)
				}
				if matches {
					return p.ID, nil
				}
			}
		}
	}

	plan := domain.NewPlan(customerID, invoiceDate, &invoiceDate, int(domain.PlanTypeManual), nil)
	if err := s.repo.SavePlan(plan); err != nil {
		return 0, err
	}
	return plan.ID, nil
}

func (s *InvoiceCreate) mapInvoiceToDTO(inv *domain.Invoice, customerNickname string) *dto.InvoiceDTO {
	if inv == nil {
		return nil
	}
	itemsDTO := make([]dto.InvoiceItemDTO, 0, len(inv.Items))
	for _, it := range inv.Items {
		desc := ""
		if it.Description != nil {
			desc = *it.Description
		}
		itemsDTO = append(itemsDTO, dto.InvoiceItemDTO{
			ID:               it.ID,
			ServiceID:        it.ServiceID,
			Description:      desc,
			Price:            it.Price,
			InvoicedQuantity: it.Quantity,
			RealizedQuantity: it.Quantity,
		})
	}

	return &dto.InvoiceDTO{
		ID:                 inv.ID,
		InvoiceDate:        inv.InvoicingDate.Format("2006-01-02"),
		DueDate:            inv.DueDate.Format("2006-01-02"),
		PeriodStart:        inv.PeriodStart.Format("2006-01-02"),
		PeriodEnd:          inv.PeriodEnd.Format("2006-01-02"),
		Nickname:           customerNickname,
		Amount:             inv.Amount,
		ReferenceInvoiceID: inv.ReferenceInvoiceID,
		Invoiced:           inv.ReferenceInvoiceID != nil,
		Items:              itemsDTO,
	}
}
