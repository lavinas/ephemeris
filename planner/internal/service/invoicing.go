package service

import (
	"fmt"
	"time"

	"planner/internal/domain"
	domainInvoicing "planner/internal/domain/invoicing"
	"planner/internal/dto"
	"planner/internal/port"
)

// Invoicing is a service that handles generating invoices based on registered plans.
type Invoicing struct {
	Base
	agendaStrat     *domainInvoicing.AgendaStrategy
	packageStrat    *domainInvoicing.PackageStrategy
	notebookStrat   *domainInvoicing.NotebookStrategy
	payPerViewStrat *domainInvoicing.PayPerViewStrategy
}

// NewInvoicing creates a new instance of Invoicing service.
func NewInvoicing(repo port.Repository, logger port.Logger) *Invoicing {
	return &Invoicing{
		Base:            *NewBase(repo, logger),
		agendaStrat:     domainInvoicing.NewAgendaStrategy(),
		packageStrat:    domainInvoicing.NewPackageStrategy(),
		notebookStrat:   domainInvoicing.NewNotebookStrategy(),
		payPerViewStrat: domainInvoicing.NewPayPerViewStrategy(),
	}
}

// Run executes the invoicing process using the provided input DTO.
func (s *Invoicing) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing invoicing request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Invoicing validation error: %v", err)
		return dto.NewInvoicingResponse(400, "Bad Request", err.Error())
	}

	req, ok := input.(*dto.InvoicingRequest)
	if !ok {
		return dto.NewInvoicingResponse(400, "Bad Request", "payload de entrada inválido")
	}

	startDate, endDate, err := req.GetParsedDates(time.Now())
	if err != nil {
		s.logger.IPrintf(2, "Date parsing error: %v", err)
		return dto.NewInvoicingResponse(400, "Bad Request", fmt.Sprintf("erro ao processar datas: %v", err))
	}

	s.logger.IPrintf(2, "Buscando planos ativos para faturamento no período de %s a %s",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	plans, err := s.repo.FindActivePlansForInvoicing(req.PlanID, req.VendorID, req.CustomerID, req.PlanType, req.ServiceID, startDate, endDate)
	if err != nil {
		s.logger.IPrintf(2, "Error finding plans for invoicing: %v", err)
		return dto.NewInvoicingResponse(500, "Internal Server Error", fmt.Sprintf("erro ao buscar planos: %v", err))
	}

	if len(plans) == 0 {
		s.logger.IPrintf(2, "Nenhum plano ativo encontrado para os filtros informados no período")
		return dto.NewInvoicingResponse(200, "OK", "nenhum plano ativo encontrado para os filtros informados no período")
	}

	reprocess := req.IsReprocess()
	successCount := 0
	errorCount := 0
	var lastErr error

	for i := range plans {
		plan := &plans[i]
		var procErr error

		switch domain.PlanType(plan.PlanType) {
		case domain.PlanTypeAgenda:
			procErr = s.processAgenda(plan, startDate, endDate, reprocess)
		case domain.PlanTypePackage:
			procErr = s.processPackage(plan, reprocess)
		case domain.PlanTypeNotebook:
			procErr = s.processNotebook(plan, startDate, endDate, reprocess)
		case domain.PlanTypePayPerView:
			procErr = s.processPayPerView(plan, startDate, endDate, reprocess)
		default:
			procErr = fmt.Errorf("tipo de plano desconhecido ou não suportado: %d", plan.PlanType)
		}

		if procErr != nil {
			s.logger.IPrintf(2, "Erro ao processar faturamento do plano %d (tipo %d): %v", plan.ID, plan.PlanType, procErr)
			errorCount++
			lastErr = procErr
			continue
		}

		successCount++
	}

	s.logger.IPrintf(2, "Processamento de faturamento finalizado. Sucesso: %d, Erros: %d", successCount, errorCount)

	if successCount == 0 && errorCount > 0 {
		return dto.NewInvoicingResponse(500, "Internal Server Error",
			fmt.Sprintf("falha ao processar faturamento: %v", lastErr))
	}

	msg := fmt.Sprintf("faturamento processado com sucesso: %d planos processados, %d erros", successCount, errorCount)
	return dto.NewInvoicingResponse(200, "OK", msg)
}

func (s *Invoicing) processPackage(plan *domain.Plan, reprocess bool) error {
	existingInvoices, err := s.repo.FindInvoices(map[string]interface{}{
		"plan_id = ?":        plan.ID,
		"deleted_at IS NULL": nil,
	})
	if err != nil {
		return fmt.Errorf("plano %d: erro ao verificar faturas existentes: %w", plan.ID, err)
	}

	if len(existingInvoices) > 0 {
		if !reprocess {
			s.logger.IPrintf(2, "Plano pacote %d já possui fatura gerada (id: %d). Pulando processamento.",
				plan.ID, existingInvoices[0].ID)
			return nil
		}

		for _, inv := range existingInvoices {
			s.logger.IPrintf(2, "Reprocessando plano pacote %d: excluindo fatura anterior (id: %d).", plan.ID, inv.ID)
			if err := s.repo.DeleteInvoice(inv.ID); err != nil {
				return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d: %w", plan.ID, inv.ID, err)
			}
		}
	}

	now := time.Now()
	invoice, err := s.packageStrat.Calculate(plan, now)
	if err != nil {
		return err
	}

	if err := s.repo.SaveInvoice(invoice); err != nil {
		return fmt.Errorf("plano %d: erro ao salvar fatura pacote: %w", plan.ID, err)
	}

	s.logger.IPrintf(2, "Fatura gerada com sucesso para plano pacote %d (invoice id: %d, valor: %.2f)",
		plan.ID, invoice.ID, invoice.Amount)
	return nil
}

func (s *Invoicing) processAgenda(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	now := time.Now()
	currentMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endLimit := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !currentMonth.After(endLimit) {
		monthStart := currentMonth
		monthEnd := currentMonth.AddDate(0, 1, -1)
		monthLabel := currentMonth.Format("2006-01")

		if !s.agendaStrat.IsActiveInMonth(plan, monthStart) {
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
		}

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

			for _, inv := range existingInvoices {
				s.logger.IPrintf(2, "Reprocessando plano agenda %d para o mês %s: excluindo fatura anterior (id: %d).",
					plan.ID, monthLabel, inv.ID)
				if err := s.repo.DeleteInvoice(inv.ID); err != nil {
					return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d: %w", plan.ID, inv.ID, err)
				}
			}
		}

		invoice, err := s.agendaStrat.CalculateMonth(plan, monthStart, now)
		if err != nil {
			return fmt.Errorf("plano %d: erro ao calcular fatura agenda para %s: %w", plan.ID, monthLabel, err)
		}
		if invoice == nil {
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
		}

		if err := s.repo.SaveInvoice(invoice); err != nil {
			return fmt.Errorf("plano %d: erro ao salvar fatura agenda para %s: %w", plan.ID, monthLabel, err)
		}

		s.logger.IPrintf(2, "Fatura gerada com sucesso para plano agenda %d no mês %s (invoice id: %d, valor: %.2f)",
			plan.ID, monthLabel, invoice.ID, invoice.Amount)

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return nil
}

func (s *Invoicing) processNotebook(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	now := time.Now()
	currentMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endLimit := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !currentMonth.After(endLimit) {
		monthStart := currentMonth
		monthLabel := currentMonth.Format("2006-01")

		if !s.notebookStrat.IsActiveInMonth(plan, monthStart) {
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
		}

		monthEnd := currentMonth.AddDate(0, 1, -1)
		existingInvoices, err := s.repo.FindInvoices(map[string]interface{}{
			"plan_id = ?":         plan.ID,
			"deleted_at IS NULL":  nil,
			"invoicing_date >= ?": monthStart,
			"invoicing_date <= ?": monthEnd,
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

			for _, inv := range existingInvoices {
				s.logger.IPrintf(2, "Reprocessando plano caderneta %d para o mês %s: excluindo fatura anterior (id: %d).",
					plan.ID, monthLabel, inv.ID)
				if err := s.repo.DeleteInvoice(inv.ID); err != nil {
					return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d: %w", plan.ID, inv.ID, err)
				}
			}
		}

		prevMonthStart, prevMonthEnd := s.notebookStrat.PreviousMonthPeriod(monthStart)
		sessionsByService := make(map[int64][]domain.Session, len(plan.Items))
		for _, item := range plan.Items {
			sessions, err := s.repo.FindBillableSessions(plan.CustomerID, item.ServiceID, prevMonthStart, prevMonthEnd)
			if err != nil {
				return fmt.Errorf("plano %d: erro ao buscar sessões do serviço %d: %w", plan.ID, item.ServiceID, err)
			}
			sessionsByService[item.ServiceID] = sessions
		}

		invoice, err := s.notebookStrat.CalculateMonth(plan, monthStart, sessionsByService, now)
		if err != nil {
			return fmt.Errorf("plano %d: erro ao calcular fatura caderneta para %s: %w", plan.ID, monthLabel, err)
		}
		if invoice == nil {
			s.logger.IPrintf(2, "Plano caderneta %d: nenhuma fatura gerada para o mês %s. Pulando.",
				plan.ID, monthLabel)
			currentMonth = currentMonth.AddDate(0, 1, 0)
			continue
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

func (s *Invoicing) processPayPerView(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	now := time.Now()
	for _, item := range plan.Items {
		sessions, err := s.repo.FindBillableSessions(plan.CustomerID, item.ServiceID, startDate, endDate)
		if err != nil {
			return fmt.Errorf("plano %d: erro ao buscar sessões do serviço %d: %w", plan.ID, item.ServiceID, err)
		}

		for _, session := range sessions {
			existingInvoice, err := s.repo.GetInvoiceForSession(session.ID)
			if err != nil {
				return fmt.Errorf("plano %d: erro ao verificar fatura para a sessão %d: %w", plan.ID, session.ID, err)
			}

			if existingInvoice != nil {
				if !reprocess {
					s.logger.IPrintf(2, "Plano pay-per-view %d: sessão %d já faturada (invoice id: %d). Pulando sessão.",
						plan.ID, session.ID, existingInvoice.ID)
					continue
				}

				s.logger.IPrintf(2, "Reprocessando plano pay-per-view %d para sessão %d: excluindo fatura anterior (id: %d).",
					plan.ID, session.ID, existingInvoice.ID)
				if err := s.repo.DeleteInvoice(existingInvoice.ID); err != nil {
					return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d da sessão %d: %w",
						plan.ID, existingInvoice.ID, session.ID, err)
				}
			}

			invoice, err := s.payPerViewStrat.CalculateSession(plan, item, session, now)
			if err != nil {
				return fmt.Errorf("plano %d: erro ao calcular fatura pay-per-view para sessão %d: %w",
					plan.ID, session.ID, err)
			}

			if err := s.repo.SaveInvoice(invoice); err != nil {
				return fmt.Errorf("plano %d: erro ao salvar fatura pay-per-view para sessão %d: %w",
					plan.ID, session.ID, err)
			}

			s.logger.IPrintf(2, "Fatura gerada com sucesso para plano pay-per-view %d, sessão %d (invoice id: %d, valor: %.2f)",
				plan.ID, session.ID, invoice.ID, invoice.Amount)
		}
	}

	return nil
}
