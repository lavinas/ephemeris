package service

import (
	"time"

	domainInvoicing "planner/internal/domain/invoicing"
	domainReconciliation "planner/internal/domain/reconciliation"
	"planner/internal/port"
)

// Reconciliation is the application service handling reconciliation of sessions and invoices.
type Reconciliation struct {
	repo     port.ReconciliationRepository
	logger   port.Logger
	ppvStrat *domainInvoicing.PayPerViewStrategy
}

// NewReconciliation creates a new Reconciliation service instance.
func NewReconciliation(repo port.ReconciliationRepository, logger port.Logger) *Reconciliation {
	return &Reconciliation{
		repo:     repo,
		logger:   logger,
		ppvStrat: domainInvoicing.NewPayPerViewStrategy(),
	}
}

// ReconcileSession reconciles a newly created session with an active invoice item or generates a pay-per-view invoice.
func (s *Reconciliation) ReconcileSession(sessionID int64) error {
	s.logger.IPrintf(2, "Iniciando conciliação da sessão %d", sessionID)

	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		s.logger.IPrintf(2, "Erro ao buscar sessão %d para conciliação: %v", sessionID, err)
		return err
	}
	if session == nil {
		s.logger.IPrintf(2, "Sessão %d não encontrada para conciliação", sessionID)
		return nil
	}

	if !domainReconciliation.IsSessionEligible(session) {
		s.logger.IPrintf(2, "Sessão %d com status '%s' não é elegível para conciliação", sessionID, session.SessionStatus)
		return nil
	}

	isReconciled, err := s.repo.IsSessionReconciled(sessionID)
	if err != nil {
		s.logger.IPrintf(2, "Erro ao verificar se sessão %d já está conciliada: %v", sessionID, err)
		return err
	}
	if isReconciled {
		s.logger.IPrintf(2, "Sessão %d já está conciliada a uma fatura ativa", sessionID)
		return nil
	}

	item, invoice, err := s.repo.FindMatchingInvoiceItem(session.CustomerID, session.ServiceID, session.SessionDate)
	if err != nil {
		s.logger.IPrintf(2, "Erro ao buscar item de fatura correspondente para sessão %d: %v", sessionID, err)
		return err
	}

	if item != nil && invoice != nil {
		s.logger.IPrintf(2, "Encontrado item de fatura %d (fatura %d) para conciliação da sessão %d", item.ID, invoice.ID, sessionID)
		if err := s.repo.AssociateSessionToInvoiceItem(item.ID, session.ID); err != nil {
			s.logger.IPrintf(2, "Erro ao associar sessão %d ao item %d: %v", session.ID, item.ID, err)
			return err
		}
		s.logger.IPrintf(2, "Sessão %d conciliada com sucesso ao item %d da fatura %d", session.ID, item.ID, invoice.ID)
		return nil
	}

	// Se não encontrou item de fatura pré-existente, verifica se o cliente possui plano Pay-Per-View
	plan, planItem, err := s.repo.FindActivePayPerViewPlan(session.CustomerID, session.ServiceID, session.SessionDate)
	if err != nil {
		s.logger.IPrintf(2, "Erro ao verificar plano pay-per-view para sessão %d: %v", sessionID, err)
		return err
	}

	if plan != nil && planItem != nil {
		s.logger.IPrintf(2, "Cliente %d possui plano pay-per-view %d para serviço %d. Gerando pré-fatura sob demanda.",
			session.CustomerID, plan.ID, session.ServiceID)

		now := time.Now()
		ppvInvoice, err := s.ppvStrat.CalculateSession(plan, *planItem, *session, now)
		if err != nil {
			s.logger.IPrintf(2, "Erro ao calcular fatura pay-per-view para sessão %d: %v", sessionID, err)
			return err
		}

		if err := s.repo.SaveInvoice(ppvInvoice); err != nil {
			s.logger.IPrintf(2, "Erro ao salvar fatura pay-per-view para sessão %d: %v", sessionID, err)
			return err
		}

		s.logger.IPrintf(2, "Fatura pay-per-view %d gerada e conciliada com sucesso para sessão %d", ppvInvoice.ID, sessionID)
		return nil
	}

	s.logger.IPrintf(2, "Nenhum item de fatura ou plano pay-per-view encontrado para conciliar a sessão %d (aguardará faturamento futuro)", sessionID)
	return nil
}

// DeconcileSession removes the link of a deleted session and decrements realized_quantity on the invoice item.
func (s *Reconciliation) DeconcileSession(sessionID int64) error {
	s.logger.IPrintf(2, "Iniciando desconciliação da sessão %d", sessionID)
	item, err := s.repo.DissociateSessionFromInvoiceItem(sessionID)
	if err != nil {
		s.logger.IPrintf(2, "Erro ao desconciliar sessão %d: %v", sessionID, err)
		return err
	}
	if item != nil {
		s.logger.IPrintf(2, "Sessão %d desconciliada do item %d com sucesso (nova quant. realizada: %d)",
			sessionID, item.ID, item.RealizedQuantity)
	} else {
		s.logger.IPrintf(2, "Sessão %d não possuía associação ativa para desconciliar", sessionID)
	}
	return nil
}

// ReconcileInvoice reconciles an invoice with all existing matching unreconciled sessions within its period.
func (s *Reconciliation) ReconcileInvoice(invoiceID int64) error {
	s.logger.IPrintf(2, "Iniciando conciliação da fatura %d", invoiceID)
	invoice, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Erro ao buscar fatura %d para conciliação: %v", invoiceID, err)
		return err
	}
	if invoice == nil {
		s.logger.IPrintf(2, "Fatura %d não encontrada para conciliação", invoiceID)
		return nil
	}
	if invoice.Plan == nil {
		s.logger.IPrintf(2, "Fatura %d sem plano associado", invoiceID)
		return nil
	}

	customerID := invoice.Plan.CustomerID
	reconciledCount := 0

	for _, item := range invoice.Items {
		unreconciled, err := s.repo.FindUnreconciledSessions(customerID, item.ServiceID, invoice.PeriodStart, invoice.PeriodEnd)
		if err != nil {
			s.logger.IPrintf(2, "Erro ao buscar sessões não conciliadas para item %d (serviço %d): %v", item.ID, item.ServiceID, err)
			return err
		}

		for _, session := range unreconciled {
			if err := s.repo.AssociateSessionToInvoiceItem(item.ID, session.ID); err != nil {
				s.logger.IPrintf(2, "Erro ao associar sessão %d ao item %d na conciliação de fatura: %v", session.ID, item.ID, err)
				return err
			}
			reconciledCount++
		}
	}

	s.logger.IPrintf(2, "Fatura %d conciliada com sucesso (%d sessões associadas)", invoiceID, reconciledCount)
	return nil
}
