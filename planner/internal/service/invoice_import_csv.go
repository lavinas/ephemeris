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

// InvoiceImportCSV is a service that handles importing invoices from a CSV file.
type InvoiceImportCSV struct {
	Base
	comptime port.ComptimeService
}

// NewInvoiceImportCSV creates a new instance of InvoiceImportCSV service.
func NewInvoiceImportCSV(repo port.Repository, logger port.Logger) *InvoiceImportCSV {
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &InvoiceImportCSV{
		Base:     *NewBase(repo, logger),
		comptime: comptime,
	}
}

// SetComptimeService sets the comptime service.
func (s *InvoiceImportCSV) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}

// Run executes the CSV import process.
func (s *InvoiceImportCSV) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing CSV invoice import request")
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "CSV invoice import validation error: %v", err)
		return dto.NewInvoiceImportCSVResponse(400, "Bad Request", err.Error(), 0, 0, 0)
	}

	req, ok := input.(*dto.InvoiceImportCSVRequest)
	if !ok {
		return dto.NewInvoiceImportCSVResponse(400, "Bad Request", "payload de entrada inválido", 0, 0, 0)
	}

	if len(req.GroupedInvoices) == 0 {
		return dto.NewInvoiceImportCSVResponse(400, "Bad Request", "nenhuma fatura para importar", 0, 0, 0)
	}

	invoicesToSave := make([]*domain.Invoice, 0, len(req.GroupedInvoices))
	totalItems := 0
	totalAmount := 0.0

	for _, group := range req.GroupedInvoices {
		serviceIDs := make([]int64, 0, len(group.Items))
		for _, it := range group.Items {
			serviceIDs = append(serviceIDs, it.ServiceID)
		}

		planID, err := s.resolveOrCreatePlan(group.CustomerID, serviceIDs, group.PeriodStart, group.Items)
		if err != nil {
			s.logger.IPrintf(2, "Failed to resolve plan for customer %s (ID %d): %v", group.CustomerNickname, group.CustomerID, err)
			return dto.NewInvoiceImportCSVResponse(500, "error", fmt.Sprintf("linha %d: erro ao associar plano: %v", group.FirstLineNumber, err), 0, 0, 0)
		}

		invTotal := 0.0
		items := make([]domain.InvoiceItem, len(group.Items))
		for i, it := range group.Items {
			lineTotal := it.Price * float64(it.Quantity)
			invTotal += lineTotal
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

		inv := domain.NewInvoice(planID, group.InvoiceDate, group.DueDate, group.PeriodStart, group.PeriodEnd, invTotal, group.ReferenceInvoiceID)
		inv.Items = items

		invoicesToSave = append(invoicesToSave, inv)
		totalItems += len(items)
		totalAmount += invTotal
	}

	// Persist all invoices and their items atomically in a single transaction
	if err := s.repo.SaveInvoices(invoicesToSave); err != nil {
		s.logger.IPrintf(2, "Failed to save imported invoices: %v", err)
		return dto.NewInvoiceImportCSVResponse(500, "error", fmt.Sprintf("erro ao persistir faturas na base de dados: %v", err), 0, 0, 0)
	}

	// Trigger comptime processing for created invoice items
	if s.comptime != nil {
		for _, inv := range invoicesToSave {
			for _, it := range inv.Items {
				if err := s.comptime.ProcessInvoiceItemCreated(context.Background(), it.ID); err != nil {
					s.logger.IPrintf(2, "Aviso: falha ao processar item importado %d no banco de horas: %v", it.ID, err)
				}
			}
		}
	}

	s.logger.IPrintf(2, "CSV import completed successfully: %d invoices, %d items, total %.2f", len(invoicesToSave), totalItems, totalAmount)

	msg := fmt.Sprintf("%d faturas e %d itens importados com sucesso", len(invoicesToSave), totalItems)
	return dto.NewInvoiceImportCSVResponse(200, "success", msg, len(invoicesToSave), totalItems, totalAmount)
}

func (s *InvoiceImportCSV) resolveOrCreatePlan(customerID int64, serviceIDs []int64, periodStart time.Time, items []dto.CSVInvoiceItem) (int64, error) {
	plans, err := s.repo.FindCustomerActivePlansWithServices(customerID, serviceIDs, 0)
	if err == nil && len(plans) > 0 {
		return plans[0].ID, nil
	}

	activePlans, _, err := s.repo.FindPlans(1, 1, map[string]interface{}{
		"planner.plan.customer_id = ?": customerID,
		"planner.plan.deleted_at IS NULL": nil,
	})
	if err == nil && len(activePlans) > 0 {
		return activePlans[0].ID, nil
	}

	plan := domain.NewPlan(customerID, periodStart, nil, int(domain.PlanTypePayPerView), nil)
	plan.PayPerView = &domain.PlanPayPerView{PaymentTerm: 3}
	for i, it := range items {
		p := it.Price
		plan.Items = append(plan.Items, domain.PlanItem{
			ServiceID:  it.ServiceID,
			OrderIndex: i + 1,
			Price:      &p,
		})
	}

	if err := s.repo.SavePlan(plan); err != nil {
		return 0, err
	}
	return plan.ID, nil
}
