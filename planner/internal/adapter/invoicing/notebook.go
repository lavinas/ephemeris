package invoicing

import (
	"errors"
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// NotebookStrategy handles invoicing for notebook plans (plan_type = 3).
type NotebookStrategy struct {
	repo   port.Repository
	logger port.Logger
}

// NewNotebookStrategy creates a new NotebookStrategy adapter.
func NewNotebookStrategy(repo port.Repository, logger port.Logger) *NotebookStrategy {
	return &NotebookStrategy{
		repo:   repo,
		logger: logger,
	}
}

// Process executes the notebook invoicing strategy.
func (s *NotebookStrategy) Process(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	if plan == nil {
		return errors.New("plano nulo")
	}
	if len(plan.Items) == 0 {
		return fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	currentMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endLimit := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !currentMonth.After(endLimit) {
		monthStart := currentMonth
		monthEnd := currentMonth.AddDate(0, 1, -1)
		monthLabel := currentMonth.Format("2006-01")

		// Verificar se o plano está ativo no mês
		isActiveInMonth := !plan.PlanStart.After(monthEnd) && (plan.PlanEnd == nil || !plan.PlanEnd.Before(monthStart))
		if !isActiveInMonth {
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
		}

		// Verificar se já existe fatura gerada para este mês para este plano
		existingInvoices, err := s.repo.FindInvoices(map[string]interface{}{
			"plan_id = ?":            plan.ID,
			"deleted_at IS NULL":     nil,
			"invoicing_date >= ?":    monthStart,
			"invoicing_date <= ?":    monthEnd,
		})
		if err != nil {
			return fmt.Errorf("plano %d: erro ao verificar faturas existentes para %s: %w", plan.ID, monthLabel, err)
		}

		if len(existingInvoices) > 0 {
			if !reprocess {
				s.logger.IPrintf(2, "Plano caderneta %d já possui fatura para o mês %s (invoice id: %d). Pulando mês.",
					plan.ID, monthLabel, existingInvoices[0].ID)
				currentMonth = currentMonth.AddDate(0, 1, 0)
				continue
			}

			// Reprocess: soft delete faturas existentes
			for _, inv := range existingInvoices {
				s.logger.IPrintf(2, "Reprocessando plano caderneta %d para o mês %s: excluindo fatura anterior (id: %d).",
					plan.ID, monthLabel, inv.ID)
				if err := s.repo.DeleteInvoice(inv.ID); err != nil {
					return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d: %w", plan.ID, inv.ID, err)
				}
			}
		}

		// Buscar sessões do mês anterior
		prevMonthStart := monthStart.AddDate(0, -1, 0)
		prevMonthEnd := monthStart.AddDate(0, 0, -1)

		invoiceItems := make([]domain.InvoiceItem, 0, len(plan.Items))
		var itemsTotalAmount float64
		var totalSessionsCount int

		for _, item := range plan.Items {
			sessions, err := s.repo.FindBillableSessions(plan.CustomerID, item.ServiceID, prevMonthStart, prevMonthEnd)
			if err != nil {
				return fmt.Errorf("plano %d: erro ao buscar sessões do serviço %d: %w", plan.ID, item.ServiceID, err)
			}

			invoicedQty := len(sessions)
			totalSessionsCount += invoicedQty

			price := 0.0
			if item.Price != nil {
				price = *item.Price
			}

			var desc *string
			if item.Service != nil && item.Service.Description != nil {
				descCopy := *item.Service.Description
				desc = &descCopy
			}

			itemSessions := make([]domain.InvoiceItemSession, invoicedQty)
			for idx, sess := range sessions {
				itemSessions[idx] = domain.InvoiceItemSession{
					SessionID: sess.ID,
					CreatedAt: now,
				}
			}

			invItem := domain.InvoiceItem{
				ServiceID:        item.ServiceID,
				Price:            price,
				Description:      desc,
				InvoicedQuantity: invoicedQty,
				RealizedQuantity: invoicedQty,
				CreatedAt:        now,
				UpdatedAt:        now,
				DeletedAt:        nil,
				ItemSessions:     itemSessions,
			}

			itemsTotalAmount += price * float64(invoicedQty)
			invoiceItems = append(invoiceItems, invItem)
		}

		// Se não há sessões e o plano não tem preço fixo, pula criação
		if totalSessionsCount == 0 && (plan.Price == nil || *plan.Price == 0) {
			s.logger.IPrintf(2, "Plano caderneta %d: nenhuma sessão encontrada no mês anterior (%s). Pulando faturamento do mês %s.",
				plan.ID, prevMonthStart.Format("2006-01"), monthLabel)
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
		}

		// Definir vencimento
		dueDate := currentDate
		if plan.Package != nil {
			dueDate = plan.Package.PaymentDate
			if dueDate.Before(currentDate) {
				dueDate = currentDate
			}
		} else if plan.Notebook != nil && plan.Notebook.PaymentDay > 0 {
			daysInMonth := daysInMonth(monthStart.Year(), monthStart.Month())
			payDay := plan.Notebook.PaymentDay
			if payDay > daysInMonth {
				payDay = daysInMonth
			}
			dueDate = time.Date(monthStart.Year(), monthStart.Month(), payDay, 0, 0, 0, 0, time.UTC)
			if dueDate.Before(currentDate) {
				dueDate = currentDate
			}
		}

		invoicingDate := plan.PlanStart
		if monthStart.After(plan.PlanStart) {
			invoicingDate = monthStart
		}

		amount := itemsTotalAmount
		if plan.Price != nil {
			amount = *plan.Price
		}

		invoice := &domain.Invoice{
			PlanID:             plan.ID,
			InvoicingDate:      invoicingDate,
			DueDate:            dueDate,
			Amount:             amount,
			ReferenceInvoiceID: nil,
			CreatedAt:          now,
			UpdatedAt:          now,
			DeletedAt:          nil,
			Items:              invoiceItems,
		}

		if err := s.repo.SaveInvoice(invoice); err != nil {
			return fmt.Errorf("plano %d: erro ao salvar fatura caderneta para %s: %w", plan.ID, monthLabel, err)
		}

		s.logger.IPrintf(2, "Fatura gerada com sucesso para plano caderneta %d no mês %s (invoice id: %d, valor: %.2f)",
			plan.ID, monthLabel, invoice.ID, invoice.Amount)

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return nil
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
