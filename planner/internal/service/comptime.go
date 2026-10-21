package service

import (
	"context"
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// ComptimeService implements port.ComptimeService for managing banco de horas.
type ComptimeService struct {
	repo   port.ComptimeRepository
	logger port.Logger
}

// NewComptime creates a new ComptimeService instance.
func NewComptime(repo port.ComptimeRepository, logger port.Logger) *ComptimeService {
	return &ComptimeService{
		repo:   repo,
		logger: logger,
	}
}

// recalculateAndPropagate calculates the balance of the given comptime record
// based on previous month's balance, saves it, and propagates the new balance
// in cascade to all subsequent months for the same customer and service.
func (s *ComptimeService) recalculateAndPropagate(ctx context.Context, comptime *domain.Comptime) error {
	prev, err := s.repo.GetPreviousComptime(ctx, comptime.CustomerID, comptime.ServiceID, comptime.Month)
	if err != nil {
		return fmt.Errorf("erro ao buscar saldo anterior para cliente %d serviço %d mês %v: %w",
			comptime.CustomerID, comptime.ServiceID, comptime.Month, err)
	}
	prevBalance := 0
	if prev != nil {
		prevBalance = prev.Balance
	}

	comptime.RecalculateBalance(prevBalance)
	if err := s.repo.SaveComptime(ctx, comptime); err != nil {
		return fmt.Errorf("erro ao salvar comptime id %d: %w", comptime.ID, err)
	}

	// Propaga para os meses subsequentes ordenados cronologicamente
	currentBalance := comptime.Balance
	subsequents, err := s.repo.GetSubsequentComptimes(ctx, comptime.CustomerID, comptime.ServiceID, comptime.Month)
	if err != nil {
		return fmt.Errorf("erro ao buscar meses subsequentes para cliente %d serviço %d: %w",
			comptime.CustomerID, comptime.ServiceID, err)
	}

	for i := range subsequents {
		subsequents[i].RecalculateBalance(currentBalance)
		if err := s.repo.SaveComptime(ctx, &subsequents[i]); err != nil {
			return fmt.Errorf("erro ao propagar saldo subsequente id %d: %w", subsequents[i].ID, err)
		}
		currentBalance = subsequents[i].Balance
	}

	return nil
}

// ProcessSessionCreated accounts for a newly created session in banco de horas.
func (s *ComptimeService) ProcessSessionCreated(ctx context.Context, sessionID int64) error {
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar sessão %d: %w", sessionID, err)
	}
	if session == nil || session.DeletedAt != nil {
		return nil
	}

	// Lock transacional para serializar mutações do mesmo cliente e serviço
	_ = s.repo.LockCustomerService(ctx, session.CustomerID, session.ServiceID)

	// Idempotência: verifica se a sessão já está registrada em comptime_session ativo
	existingCS, err := s.repo.FindComptimeSessionBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("erro ao verificar comptime_session para sessão %d: %w", sessionID, err)
	}
	if existingCS != nil {
		return nil // já contabilizada
	}

	vendorID := int64(1)
	if session.Customer != nil && session.Customer.VendorID > 0 {
		vendorID = session.Customer.VendorID
	}

	effectiveMinutes := domain.EffectiveSessionMinutes(session.SessionMinutes, session.SessionStatus)
	month := domain.NormalizeMonth(session.SessionDate)

	comptime, err := s.repo.FindOrCreateComptime(ctx, vendorID, session.CustomerID, session.ServiceID, month)
	if err != nil {
		return fmt.Errorf("erro ao obter ou criar comptime para cliente %d serviço %d mês %v: %w",
			session.CustomerID, session.ServiceID, month, err)
	}

	comptime.ConsumedMinutes += effectiveMinutes

	cs := domain.NewComptimeSession(comptime.ID, sessionID, month, effectiveMinutes)
	if err := s.repo.SaveComptimeSession(ctx, cs); err != nil {
		return fmt.Errorf("erro ao salvar comptime_session para sessão %d: %w", sessionID, err)
	}

	if err := s.recalculateAndPropagate(ctx, comptime); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.IPrintf(2, "Comptime: sessão %d contabilizada com %d minutos no mês %s",
			sessionID, effectiveMinutes, month.Format("2006-01"))
	}
	return nil
}

// ProcessSessionDeleted accounts for a deleted session by subtracting its minutes and soft-deleting comptime_session.
func (s *ComptimeService) ProcessSessionDeleted(ctx context.Context, sessionID int64) error {
	cs, err := s.repo.FindComptimeSessionBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime_session para sessão %d: %w", sessionID, err)
	}
	if cs == nil {
		return nil
	}

	comptime, err := s.repo.GetComptime(ctx, cs.ComptimeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", cs.ComptimeID, err)
	}
	if comptime == nil {
		return nil
	}

	_ = s.repo.LockCustomerService(ctx, comptime.CustomerID, comptime.ServiceID)

	comptime.ConsumedMinutes -= cs.ConsumedMinutes
	if comptime.ConsumedMinutes < 0 {
		comptime.ConsumedMinutes = 0
	}

	if err := s.repo.DeleteComptimeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("erro ao deletar comptime_session para sessão %d: %w", sessionID, err)
	}

	if err := s.recalculateAndPropagate(ctx, comptime); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.IPrintf(2, "Comptime: sessão %d removida do banco de horas (estorno de %d min)",
			sessionID, cs.ConsumedMinutes)
	}
	return nil
}

// ProcessSessionUpdated recalculates comptime upon session edits.
func (s *ComptimeService) ProcessSessionUpdated(ctx context.Context, sessionID int64) error {
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar sessão %d: %w", sessionID, err)
	}
	if session == nil {
		return nil
	}
	if session.DeletedAt != nil {
		return s.ProcessSessionDeleted(ctx, sessionID)
	}

	cs, err := s.repo.FindComptimeSessionBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime_session para sessão %d: %w", sessionID, err)
	}
	if cs == nil {
		return s.ProcessSessionCreated(ctx, sessionID)
	}

	_ = s.repo.LockCustomerService(ctx, session.CustomerID, session.ServiceID)

	newEffective := domain.EffectiveSessionMinutes(session.SessionMinutes, session.SessionStatus)
	newMonth := domain.NormalizeMonth(session.SessionDate)

	comptime, err := s.repo.GetComptime(ctx, cs.ComptimeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", cs.ComptimeID, err)
	}
	if comptime == nil {
		return s.ProcessSessionCreated(ctx, sessionID)
	}

	// Caso o mês, cliente ou serviço tenha mudado, move a sessão entre registros de comptime
	if !newMonth.Equal(cs.Month) || session.CustomerID != comptime.CustomerID || session.ServiceID != comptime.ServiceID {
		// Estorna da antiga
		comptime.ConsumedMinutes -= cs.ConsumedMinutes
		if comptime.ConsumedMinutes < 0 {
			comptime.ConsumedMinutes = 0
		}
		_ = s.repo.DeleteComptimeSession(ctx, sessionID)
		_ = s.recalculateAndPropagate(ctx, comptime)

		// Cria no novo comptime
		return s.ProcessSessionCreated(ctx, sessionID)
	}

	// Mesmo mês e serviço: atualiza a diferença
	delta := newEffective - cs.ConsumedMinutes
	comptime.ConsumedMinutes += delta
	if comptime.ConsumedMinutes < 0 {
		comptime.ConsumedMinutes = 0
	}

	cs.ConsumedMinutes = newEffective
	if err := s.repo.SaveComptimeSession(ctx, cs); err != nil {
		return fmt.Errorf("erro ao atualizar comptime_session da sessão %d: %w", sessionID, err)
	}

	return s.recalculateAndPropagate(ctx, comptime)
}

// ProcessInvoiceItemCreated accounts for a new invoice item in banco de horas.
func (s *ComptimeService) ProcessInvoiceItemCreated(ctx context.Context, invoiceItemID int64) error {
	existingCII, err := s.repo.FindComptimeInvoiceItemByInvoiceItemID(ctx, invoiceItemID)
	if err != nil {
		return fmt.Errorf("erro ao verificar comptime_invoice_item para item %d: %w", invoiceItemID, err)
	}
	if existingCII != nil {
		return nil // já contabilizada
	}

	item, err := s.repo.GetInvoiceItemByID(ctx, invoiceItemID)
	if err != nil {
		return fmt.Errorf("erro ao buscar invoice_item %d: %w", invoiceItemID, err)
	}
	if item == nil || item.DeletedAt != nil || item.Invoice == nil || item.Invoice.Plan == nil {
		return nil
	}

	customerID := item.Invoice.Plan.CustomerID
	serviceID := item.ServiceID
	vendorID := int64(1)
	if item.Invoice.Plan.Customer != nil && item.Invoice.Plan.Customer.VendorID > 0 {
		vendorID = item.Invoice.Plan.Customer.VendorID
	}

	_ = s.repo.LockCustomerService(ctx, customerID, serviceID)

	sessionMinutes := 60
	if item.Service != nil && item.Service.SessionMinutes != nil && *item.Service.SessionMinutes > 0 {
		sessionMinutes = *item.Service.SessionMinutes
	}

	invoicedMinutes := item.Quantity * sessionMinutes
	month := domain.NormalizeMonth(item.Invoice.PeriodStart)

	comptime, err := s.repo.FindOrCreateComptime(ctx, vendorID, customerID, serviceID, month)
	if err != nil {
		return fmt.Errorf("erro ao obter ou criar comptime para fatura item %d: %w", invoiceItemID, err)
	}

	comptime.InvoicedMinutes += invoicedMinutes

	cii := domain.NewComptimeInvoiceItem(comptime.ID, invoiceItemID, month, invoicedMinutes)
	if err := s.repo.SaveComptimeInvoiceItem(ctx, cii); err != nil {
		return fmt.Errorf("erro ao salvar comptime_invoice_item: %w", err)
	}

	if err := s.recalculateAndPropagate(ctx, comptime); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.IPrintf(2, "Comptime: item fatura %d creditado com %d minutos no mês %s",
			invoiceItemID, invoicedMinutes, month.Format("2006-01"))
	}
	return nil
}

// ProcessInvoiceItemDeleted accounts for a deleted invoice item.
func (s *ComptimeService) ProcessInvoiceItemDeleted(ctx context.Context, invoiceItemID int64) error {
	cii, err := s.repo.FindComptimeInvoiceItemByInvoiceItemID(ctx, invoiceItemID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime_invoice_item para item %d: %w", invoiceItemID, err)
	}
	if cii == nil {
		return nil
	}

	comptime, err := s.repo.GetComptime(ctx, cii.ComptimeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", cii.ComptimeID, err)
	}
	if comptime == nil {
		return nil
	}

	_ = s.repo.LockCustomerService(ctx, comptime.CustomerID, comptime.ServiceID)

	comptime.InvoicedMinutes -= cii.InvoicedMinutes
	if comptime.InvoicedMinutes < 0 {
		comptime.InvoicedMinutes = 0
	}

	if err := s.repo.DeleteComptimeInvoiceItem(ctx, invoiceItemID); err != nil {
		return fmt.Errorf("erro ao deletar comptime_invoice_item %d: %w", invoiceItemID, err)
	}

	if err := s.recalculateAndPropagate(ctx, comptime); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.IPrintf(2, "Comptime: item fatura %d estornado com %d minutos",
			invoiceItemID, cii.InvoicedMinutes)
	}
	return nil
}

// ProcessInvoiceItemUpdated accounts for an updated invoice item.
func (s *ComptimeService) ProcessInvoiceItemUpdated(ctx context.Context, invoiceItemID int64) error {
	item, err := s.repo.GetInvoiceItemByID(ctx, invoiceItemID)
	if err != nil {
		return fmt.Errorf("erro ao buscar item %d: %w", invoiceItemID, err)
	}
	if item == nil {
		return nil
	}
	if item.DeletedAt != nil {
		return s.ProcessInvoiceItemDeleted(ctx, invoiceItemID)
	}

	cii, err := s.repo.FindComptimeInvoiceItemByInvoiceItemID(ctx, invoiceItemID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime_invoice_item para item %d: %w", invoiceItemID, err)
	}
	if cii == nil {
		return s.ProcessInvoiceItemCreated(ctx, invoiceItemID)
	}

	comptime, err := s.repo.GetComptime(ctx, cii.ComptimeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", cii.ComptimeID, err)
	}
	if comptime == nil {
		return s.ProcessInvoiceItemCreated(ctx, invoiceItemID)
	}

	_ = s.repo.LockCustomerService(ctx, comptime.CustomerID, comptime.ServiceID)

	sessionMinutes := 60
	if item.Service != nil && item.Service.SessionMinutes != nil && *item.Service.SessionMinutes > 0 {
		sessionMinutes = *item.Service.SessionMinutes
	}

	newInvoicedMinutes := item.Quantity * sessionMinutes
	delta := newInvoicedMinutes - cii.InvoicedMinutes

	comptime.InvoicedMinutes += delta
	if comptime.InvoicedMinutes < 0 {
		comptime.InvoicedMinutes = 0
	}

	cii.InvoicedMinutes = newInvoicedMinutes
	if err := s.repo.SaveComptimeInvoiceItem(ctx, cii); err != nil {
		return fmt.Errorf("erro ao salvar comptime_invoice_item %d: %w", invoiceItemID, err)
	}

	return s.recalculateAndPropagate(ctx, comptime)
}

// ProcessInvoiceDeleted accounts for an entire deleted invoice.
func (s *ComptimeService) ProcessInvoiceDeleted(ctx context.Context, invoiceID int64) error {
	items, err := s.repo.FindComptimeInvoiceItemsByInvoiceID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("erro ao buscar itens de comptime para fatura %d: %w", invoiceID, err)
	}

	for _, it := range items {
		if err := s.ProcessInvoiceItemDeleted(ctx, it.InvoiceItemID); err != nil {
			if s.logger != nil {
				s.logger.IPrintf(1, "Aviso: falha ao remover item de fatura %d do comptime: %v", it.InvoiceItemID, err)
			}
		}
	}
	return nil
}

// ProcessComptimeExtraCreated adds extra minutes to comptime and recalculates balance.
func (s *ComptimeService) ProcessComptimeExtraCreated(ctx context.Context, extra *domain.ComptimeExtra) error {
	comptime, err := s.repo.GetComptime(ctx, extra.ComptimeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", extra.ComptimeID, err)
	}
	if comptime == nil {
		return fmt.Errorf("comptime %d não encontrado", extra.ComptimeID)
	}

	_ = s.repo.LockCustomerService(ctx, comptime.CustomerID, comptime.ServiceID)

	comptime.InvoicedMinutes += extra.InvoicedMinutes
	comptime.ConsumedMinutes += extra.ConsumedMinutes

	if err := s.repo.SaveComptimeExtra(ctx, extra); err != nil {
		return fmt.Errorf("erro ao salvar comptime_extra: %w", err)
	}

	return s.recalculateAndPropagate(ctx, comptime)
}

// ProcessComptimeExtraUpdated updates an existing comptime_extra adjustment.
func (s *ComptimeService) ProcessComptimeExtraUpdated(ctx context.Context, extra *domain.ComptimeExtra) error {
	comptime, err := s.repo.GetComptime(ctx, extra.ComptimeID)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", extra.ComptimeID, err)
	}
	if comptime == nil {
		return fmt.Errorf("comptime %d não encontrado", extra.ComptimeID)
	}

	_ = s.repo.LockCustomerService(ctx, comptime.CustomerID, comptime.ServiceID)

	if err := s.repo.SaveComptimeExtra(ctx, extra); err != nil {
		return fmt.Errorf("erro ao atualizar comptime_extra: %w", err)
	}

	return s.recalculateAndPropagate(ctx, comptime)
}

// ProcessComptimeExtraDeleted soft deletes extra adjustment and recalculates.
func (s *ComptimeService) ProcessComptimeExtraDeleted(ctx context.Context, extraID int64) error {
	return s.repo.DeleteComptimeExtra(ctx, extraID)
}

// Resync executes full resynchronization of sessions and invoices within the given date period,
// optionally filtered by customer and service. If not specified (nil or <= 0), it syncs all.
func (s *ComptimeService) Resync(ctx context.Context, vendorID int64, startDate, endDate time.Time, dataType string, customerID, serviceID *int64) (*port.ResyncResult, error) {
	startExec := time.Now()
	if dataType == "" || dataType == "todos" {
		dataType = "all"
	}
	if customerID != nil && *customerID <= 0 {
		customerID = nil
	}
	if serviceID != nil && *serviceID <= 0 {
		serviceID = nil
	}

	result := &port.ResyncResult{
		VendorID:   vendorID,
		CustomerID: customerID,
		ServiceID:  serviceID,
		StartDate:  startDate,
		EndDate:    endDate,
		DataType:   dataType,
	}

	affectedComptimes := make(map[string]bool)

	if dataType == "session" || dataType == "all" {
		sessions, err := s.repo.FindActiveSessionsForComptime(ctx, vendorID, startDate, endDate, customerID, serviceID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar sessões para ressincronização: %w", err)
		}
		for _, sess := range sessions {
			if err := s.ProcessSessionCreated(ctx, sess.ID); err != nil {
				if s.logger != nil {
					s.logger.IPrintf(1, "Aviso na ressincronização da sessão %d: %v", sess.ID, err)
				}
			} else {
				result.SessionsSynced++
				key := fmt.Sprintf("%d_%d_%s", sess.CustomerID, sess.ServiceID, domain.NormalizeMonth(sess.SessionDate).Format("2006-01"))
				affectedComptimes[key] = true
			}
		}
	}

	if dataType == "invoice" || dataType == "all" {
		items, err := s.repo.FindActiveInvoiceItemsForComptime(ctx, vendorID, startDate, endDate, customerID, serviceID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar faturas para ressincronização: %w", err)
		}
		for _, it := range items {
			if err := s.ProcessInvoiceItemCreated(ctx, it.ID); err != nil {
				if s.logger != nil {
					s.logger.IPrintf(1, "Aviso na ressincronização do item de fatura %d: %v", it.ID, err)
				}
			} else {
				result.InvoicesSynced++
				cID := int64(0)
				if it.Invoice != nil && it.Invoice.Plan != nil {
					cID = it.Invoice.Plan.CustomerID
				}
				pStart := time.Time{}
				if it.Invoice != nil {
					pStart = it.Invoice.PeriodStart
				}
				key := fmt.Sprintf("%d_%d_%s", cID, it.ServiceID, domain.NormalizeMonth(pStart).Format("2006-01"))
				affectedComptimes[key] = true
			}
		}
	}

	result.ComptimeUpdated = len(affectedComptimes)
	result.Duration = time.Since(startExec)
	return result, nil
}

// DeleteComptime soft-deletes a comptime record, its associations, and recalculates subsequent months' balances.
func (s *ComptimeService) DeleteComptime(ctx context.Context, id int64) error {
	comptime, err := s.repo.GetComptime(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao buscar comptime id %d: %w", id, err)
	}
	if comptime == nil {
		return fmt.Errorf("registro de banco de horas não encontrado")
	}

	_ = s.repo.LockCustomerService(ctx, comptime.CustomerID, comptime.ServiceID)

	if err := s.repo.DeleteComptime(ctx, id); err != nil {
		return fmt.Errorf("erro ao excluir banco de horas: %w", err)
	}

	// Recalcula cascata para os meses subsequentes se houver
	subsequents, err := s.repo.GetSubsequentComptimes(ctx, comptime.CustomerID, comptime.ServiceID, comptime.Month)
	if err == nil && len(subsequents) > 0 {
		_ = s.recalculateAndPropagate(ctx, &subsequents[0])
	}

	if s.logger != nil {
		s.logger.IPrintf(2, "Comptime: registro %d excluído para cliente %d serviço %d mês %s",
			id, comptime.CustomerID, comptime.ServiceID, comptime.Month.Format("2006-01"))
	}
	return nil
}

// GetComptime retrieves a comptime record by ID.
func (s *ComptimeService) GetComptime(ctx context.Context, id int64) (*domain.Comptime, error) {
	return s.repo.GetComptime(ctx, id)
}

// GetComptimeByCustomerServiceMonth retrieves a comptime record for a customer, service and month.
func (s *ComptimeService) GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	return s.repo.GetComptimeByCustomerServiceMonth(ctx, customerID, serviceID, month)
}

// ListComptime lists enriched comptime views according to the filter.
func (s *ComptimeService) ListComptime(ctx context.Context, filter port.ComptimeFilter) ([]port.ComptimeView, int64, error) {
	return s.repo.ListComptime(ctx, filter)
}
