package invoicing

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// AgendaStrategy handles invoicing for recurring agenda plans (plan_type = 1).
type AgendaStrategy struct {
	repo   port.Repository
	logger port.Logger
}

// NewAgendaStrategy creates a new AgendaStrategy adapter.
func NewAgendaStrategy(repo port.Repository, logger port.Logger) *AgendaStrategy {
	return &AgendaStrategy{
		repo:   repo,
		logger: logger,
	}
}

// Process executes the agenda invoicing strategy.
func (s *AgendaStrategy) Process(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	if plan == nil {
		return errors.New("plano nulo")
	}
	if plan.Agenda == nil {
		return fmt.Errorf("plano %d: dados de agenda não encontrados", plan.ID)
	}
	if len(plan.Items) == 0 {
		return fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	now := time.Now()
	currentMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endLimit := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !currentMonth.After(endLimit) {
		monthStart := currentMonth
		monthEnd := currentMonth.AddDate(0, 1, -1)
		monthLabel := currentMonth.Format("2006-01")

		// Faturamento feito uma vez por mês nos meses dentro do período e dentro das datas de validade do plano
		isActiveInMonth := !plan.PlanStart.After(monthEnd) && (plan.PlanEnd == nil || !plan.PlanEnd.Before(monthStart))
		if !isActiveInMonth {
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
		}

		// Verificar se já não há fatura gerada para este mês para este plano
		// A fatura do mês tem due_date no mês corrente
		existingInvoices, err := s.repo.FindInvoices(map[string]interface{}{
			"plan_id = ?":        plan.ID,
			"deleted_at IS NULL": nil,
			"due_date >= ?":      monthStart,
			"due_date <= ?":      monthEnd,
		})
		if err != nil {
			return fmt.Errorf("plano %d: erro ao verificar faturas existentes para %s: %w", plan.ID, monthLabel, err)
		}

		if len(existingInvoices) > 0 {
			if !reprocess {
				s.logger.IPrintf(2, "Plano agenda %d já possui fatura para o mês %s (invoice id: %d). Pulando mês.",
					plan.ID, monthLabel, existingInvoices[0].ID)
				currentMonth = currentMonth.AddDate(0, 1, 0)
				continue
			}

			// Reprocess: soft delete fatura existente
			for _, inv := range existingInvoices {
				s.logger.IPrintf(2, "Reprocessando plano agenda %d para o mês %s: excluindo fatura anterior (id: %d).",
					plan.ID, monthLabel, inv.ID)
				if err := s.repo.DeleteInvoice(inv.ID); err != nil {
					return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d: %w", plan.ID, inv.ID, err)
				}
			}
		}

		// Calcular período alvo de sessões de acordo com plan_agenda.term_type
		// se 1 (pre-paid): mês corrente ao processamento
		// se 2 (post-paid): mês anterior ao processamento
		targetMonth := monthStart
		if plan.Agenda.TermType == int(domain.TermTypePostPaid) {
			targetMonth = monthStart.AddDate(0, -1, 0)
		}

		targetStart := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		targetEnd := targetStart.AddDate(0, 1, -1)

		// Ajustar intervalo efetivo com vigência do plano
		effStart := targetStart
		if effStart.Before(plan.PlanStart) {
			effStart = plan.PlanStart
		}
		effEnd := targetEnd
		if plan.PlanEnd != nil && effEnd.After(*plan.PlanEnd) {
			effEnd = *plan.PlanEnd
		}

		calculatedQty := 0
		if !effStart.After(effEnd) {
			calculatedQty = s.calculateSessionsCount(plan.PlanStart, effStart, effEnd, plan.Agenda.Recurrence, plan.Agenda.WeekDay)
		}

		// Limitar por monthly_service_limit se informado
		if plan.Agenda.MonthlyServiceLimit != nil && *plan.Agenda.MonthlyServiceLimit > 0 {
			if calculatedQty > *plan.Agenda.MonthlyServiceLimit {
				calculatedQty = *plan.Agenda.MonthlyServiceLimit
			}
		}

		// Ordenar itens pelo order_index
		items := make([]domain.PlanItem, len(plan.Items))
		copy(items, plan.Items)
		sort.Slice(items, func(i, j int) bool {
			return items[i].OrderIndex < items[j].OrderIndex
		})

		nItems := len(items)
		invoiceItems := make([]domain.InvoiceItem, nItems)
		var itemsTotalAmount float64

		baseQty := calculatedQty / nItems
		remainder := calculatedQty % nItems

		for i, item := range items {
			invoicedQty := baseQty
			if i < remainder {
				invoicedQty++
			}

			price := 0.0
			if item.Price != nil {
				price = *item.Price
			}

			var desc *string
			if item.Service != nil && item.Service.Description != nil {
				descCopy := *item.Service.Description
				desc = &descCopy
			}

			invoiceItems[i] = domain.InvoiceItem{
				ServiceID:        item.ServiceID,
				Price:            price,
				Description:      desc,
				InvoicedQuantity: invoicedQty,
				RealizedQuantity: 0,
				CreatedAt:        now,
				UpdatedAt:        now,
				DeletedAt:        nil,
			}

			itemsTotalAmount += price * float64(invoicedQty)
		}

		amount := itemsTotalAmount
		if plan.Price != nil {
			amount = *plan.Price
		}

		// invoice.invoice_date = o primeiro dia do mes do periodo especificado ou a data de inicio, o que for menor
		firstDayOfMonth := monthStart
		invoicingDate := firstDayOfMonth
		if plan.PlanStart.Before(invoicingDate) {
			invoicingDate = plan.PlanStart
		}

		// invoice.due_date = mes e ano do periodo de processamento porem o dia deve-se tirar de plan_agenda.payment_day
		daysInMonth := daysInMonth(monthStart.Year(), monthStart.Month())
		payDay := plan.Agenda.PaymentDay
		if payDay > daysInMonth {
			payDay = daysInMonth
		}
		dueDate := time.Date(monthStart.Year(), monthStart.Month(), payDay, 0, 0, 0, 0, time.UTC)

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
			return fmt.Errorf("plano %d: erro ao salvar fatura agenda para %s: %w", plan.ID, monthLabel, err)
		}

		s.logger.IPrintf(2, "Fatura gerada com sucesso para plano agenda %d no mês %s (invoice id: %d, valor: %.2f, sessões calculadas: %d)",
			plan.ID, monthLabel, invoice.ID, invoice.Amount, calculatedQty)

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return nil
}

// calculateSessionsCount calculates the number of sessions according to recurrence and weekday in [effStart, effEnd]
func (s *AgendaStrategy) calculateSessionsCount(planStart, effStart, effEnd time.Time, recurrence, weekDay int) int {
	targetWeekday := time.Weekday((weekDay - 1) % 7) // 1: Sunday = 0, 2: Monday = 1 ...

	switch domain.Recurrence(recurrence) {
	case domain.RecurrenceWeekly: // 1: weekly (7 days interval)
		offset := (int(targetWeekday) - int(effStart.Weekday()) + 7) % 7
		firstOccurrence := effStart.AddDate(0, 0, offset)
		count := 0
		d := firstOccurrence
		for !d.After(effEnd) {
			count++
			d = d.AddDate(0, 0, 7)
		}
		return count

	case domain.RecurrenceBiWeekly: // 2: bi-weekly (14 days interval anchored from plan start)
		offsetPlanStart := (int(targetWeekday) - int(planStart.Weekday()) + 7) % 7
		firstPlanStart := planStart.AddDate(0, 0, offsetPlanStart)
		count := 0
		d := firstPlanStart
		for !d.After(effEnd) {
			if !d.Before(effStart) {
				count++
			}
			d = d.AddDate(0, 0, 14)
		}
		return count

	case domain.RecurrenceMonthly: // 3: monthly
		offset := (int(targetWeekday) - int(effStart.Weekday()) + 7) % 7
		firstOccurrence := effStart.AddDate(0, 0, offset)
		if !firstOccurrence.After(effEnd) {
			return 1
		}
		return 0

	default:
		// Default to weekly if unknown
		offset := (int(targetWeekday) - int(effStart.Weekday()) + 7) % 7
		firstOccurrence := effStart.AddDate(0, 0, offset)
		count := 0
		d := firstOccurrence
		for !d.After(effEnd) {
			count++
			d = d.AddDate(0, 0, 7)
		}
		return count
	}
}
