package service

import (
	"fmt"
	"time"

	"planner/internal/adapter/invoicing"
	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// Invoicing is a service that handles generating invoices based on registered plans.
type Invoicing struct {
	Base
}

// NewInvoicing creates a new instance of Invoicing service.
func NewInvoicing(repo port.Repository, logger port.Logger) *Invoicing {
	return &Invoicing{
		Base: *NewBase(repo, logger),
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

	plans, err := s.repo.FindActivePlansForInvoicing(req.VendorID, req.CustomerID, req.PlanType, req.ServiceID, startDate, endDate)
	if err != nil {
		s.logger.IPrintf(2, "Error finding plans for invoicing: %v", err)
		return dto.NewInvoicingResponse(500, "Internal Server Error", fmt.Sprintf("erro ao buscar planos: %v", err))
	}

	if len(plans) == 0 {
		s.logger.IPrintf(2, "Nenhum plano ativo encontrado para os filtros informados no período")
		return dto.NewInvoicingResponse(200, "OK", "nenhum plano ativo encontrado para os filtros informados no período")
	}

	factory := invoicing.NewStrategyFactory(s.repo, s.logger)
	reprocess := req.IsReprocess()

	successCount := 0
	errorCount := 0
	var lastErr error

	for i := range plans {
		plan := &plans[i]
		strategy, err := factory.GetStrategy(domain.PlanType(plan.PlanType))
		if err != nil {
			s.logger.IPrintf(2, "Erro ao obter estratégia para o plano %d (tipo %d): %v", plan.ID, plan.PlanType, err)
			errorCount++
			lastErr = err
			continue
		}

		if err := strategy.Process(plan, startDate, endDate, reprocess); err != nil {
			s.logger.IPrintf(2, "Erro ao processar faturamento do plano %d: %v", plan.ID, err)
			errorCount++
			lastErr = err
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
