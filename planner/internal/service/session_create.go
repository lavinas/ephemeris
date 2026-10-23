package service

import (
	"context"
	"time"

	"planner/internal/domain"
	"planner/internal/domain/invoicing"
	"planner/internal/dto"
	"planner/internal/port"
)

// SessionCreate is  a service that handles the creation of sessions.
type SessionCreate struct {
	Base
	payPerViewStrat *invoicing.PayPerViewStrategy
	comptime        port.ComptimeService
}

// NewSessionCreate creates a new SessionCreate instance.
func NewSessionCreate(repo port.Repository, logger port.Logger) *SessionCreate {
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &SessionCreate{
		Base:            *NewBase(repo, logger),
		payPerViewStrat: invoicing.NewPayPerViewStrategy(),
		comptime:        comptime,
	}
}

// SetComptimeService sets the comptime service.
func (s *SessionCreate) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}

// SetPayPerViewStrategy sets the pay-per-view strategy.
func (s *SessionCreate) SetPayPerViewStrategy(strat *invoicing.PayPerViewStrategy) {
	s.payPerViewStrat = strat
}

// Run executes the session creation process using the provided input DTO and returns an output DTO.
func (s *SessionCreate) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create session request: %v", input)
	// Validate the input DTO
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewSessionCreateResponse(400, "error", err.Error(), 0)
	}
	// Create a new session model from the input DTO
	sessionModel := s.createSessionModel(input)

	// Save the session model to the repository
	if err := s.repo.Save(sessionModel); err != nil {
		s.logger.IPrintf(2, "Failed to save session: %v", err)
		return dto.NewSessionCreateResponse(500, "error", err.Error(), 0)
	}

	session := sessionModel.(*domain.Session)
	sessionID := session.ID

	// Trigger automatic pay-per-view pre-invoicing if applicable
	s.triggerPayPerViewInvoicing(session)

	// Trigger comptime sync
	if s.comptime != nil {
		if err := s.comptime.ProcessSessionCreated(context.Background(), sessionID); err != nil {
			s.logger.IPrintf(2, "Aviso: falha na sincronização do banco de horas da sessão %d: %v", sessionID, err)
		}
	}

	// Create and return the output DTO with the session ID
	s.logger.IPrintf(2, "Successfully created session with ID: %d", sessionID)
	return dto.NewSessionCreateResponse(200, "success", "session created", sessionID)
}

// triggerPayPerViewInvoicing generates an automatic pre-invoice for a pay-per-view plan if matching.
func (s *SessionCreate) triggerPayPerViewInvoicing(session *domain.Session) {
	if session == nil || session.ID == 0 {
		return
	}
	// Se a sessão for cancelada sem cobrança, não gera pré-fatura
	if session.SessionStatus == string(domain.StatusCancelNotCharge) || session.SessionStatus == "cancelada_nao_cobrar" {
		return
	}

	ppvType := int(domain.PlanTypePayPerView)
	plans, err := s.repo.FindActivePlansForInvoicing(
		nil,
		nil,
		&session.CustomerID,
		&ppvType,
		&session.ServiceID,
		session.SessionDate,
		session.SessionDate,
	)
	if err != nil {
		s.logger.IPrintf(2, "Aviso: erro ao buscar planos pay-per-view para sessão %d: %v", session.ID, err)
		return
	}

	sessDate := time.Date(session.SessionDate.Year(), session.SessionDate.Month(), session.SessionDate.Day(), 0, 0, 0, 0, time.UTC)
	now := time.Now()

	for _, plan := range plans {
		if plan.PlanType != ppvType || plan.CustomerID != session.CustomerID {
			continue
		}

		planStart := time.Date(plan.PlanStart.Year(), plan.PlanStart.Month(), plan.PlanStart.Day(), 0, 0, 0, 0, time.UTC)
		if sessDate.Before(planStart) {
			continue
		}
		if plan.PlanEnd != nil {
			planEnd := time.Date(plan.PlanEnd.Year(), plan.PlanEnd.Month(), plan.PlanEnd.Day(), 0, 0, 0, 0, time.UTC)
			if sessDate.After(planEnd) {
				continue
			}
		}

		// Encontrar o item do plano correspondente ao serviço da sessão
		var matchingItem *domain.PlanItem
		for _, it := range plan.Items {
			if it.ServiceID == session.ServiceID {
				itemCopy := it
				matchingItem = &itemCopy
				break
			}
		}
		if matchingItem == nil {
			continue
		}

		if matchingItem.Service == nil {
			svc, err := s.repo.GetServiceByID(session.ServiceID)
			if err == nil && svc != nil {
				matchingItem.Service = svc
			}
		}

		strat := s.payPerViewStrat
		if strat == nil {
			strat = invoicing.NewPayPerViewStrategy()
		}

		invoice, err := strat.CalculateSession(&plan, *matchingItem, *session, now)
		if err != nil {
			s.logger.IPrintf(2, "Aviso: erro ao calcular pré-fatura pay-per-view para sessão %d: %v", session.ID, err)
			continue
		}

		if err := s.repo.SaveInvoice(invoice); err != nil {
			s.logger.IPrintf(2, "Aviso: erro ao salvar pré-fatura para sessão %d (plano %d): %v", session.ID, plan.ID, err)
			continue
		}

		s.logger.IPrintf(2, "Pré-fatura #%d gerada automaticamente para a sessão %d (plano pay-per-view %d)", invoice.ID, session.ID, plan.ID)

		// Ao se criar a pré-fatura, deve-se acionar o registro do banco de horas como é feito normalmente agora
		if s.comptime != nil {
			ctx := context.Background()
			for _, it := range invoice.Items {
				if err := s.comptime.ProcessInvoiceItemCreated(ctx, it.ID); err != nil {
					s.logger.IPrintf(2, "Aviso: falha ao processar item %d da pré-fatura no banco de horas: %v", it.ID, err)
				}
			}
		}

		// Apenas um plano pay-per-view faturará a sessão
		break
	}
}

// createSessionModel creates a session model from the input DTO.
func (s *SessionCreate) createSessionModel(input port.InDTO) interface{} {
	// Assuming input is of type SessionCreateRequest
	req := input.(*dto.SessionCreateRequest)
	dtime, _ := time.Parse("2006-01-02", req.Date)
	var comments *string
	if req.Comments != "" {
		comments = &req.Comments
	}
	return domain.NewSession(req.CustomerID, req.ServiceID, dtime, req.Minutes, req.Status, comments)
}
