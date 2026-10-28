package service

import (
	"context"
	"fmt"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// InvoiceTaxEmit handles the business logic for emitting a single NFS-e in real-time.
type InvoiceTaxEmit struct {
	*Base
	emitter port.TaxEmitter
	timeout time.Duration
}

// NewInvoiceTaxEmit creates a new instance of InvoiceTaxEmit.
func NewInvoiceTaxEmit(repo port.Repository, logger port.Logger, emitter port.TaxEmitter) *InvoiceTaxEmit {
	return &InvoiceTaxEmit{
		Base:    NewBase(repo, logger),
		emitter: emitter,
		timeout: 30 * time.Second,
	}
}

// Run executes the validation, external NFS-e emission call, and atomic persistence of the tax emission.
func (s *InvoiceTaxEmit) Run(inDTO port.InDTO) port.OutDTO {
	req, ok := inDTO.(*dto.InvoiceTaxEmitRequest)
	if !ok {
		s.logger.IPrintf(1, "Invalid input type for InvoiceTaxEmit: expected InvoiceTaxEmitRequest")
		return dto.NewInvoiceTaxEmitResponse(400, "error", "Tipo de entrada inválido", nil, nil, nil, 0)
	}

	s.logger.IPrintf(2, "Validating invoice tax emission request for invoice ID %d, vendor %s", req.ID, req.Vendor)
	if err := req.Validate(s.repo); err != nil {
		s.logger.IPrintf(1, "Validation failed for invoice tax emission: %v", err)
		return dto.NewInvoiceTaxEmitResponse(400, "error", fmt.Sprintf("Validação falhou: %v", err), nil, nil, nil, 0)
	}

	invoice := req.GetInvoice()
	vendor := req.GetVendor()
	if vendor != nil && (invoice.Vendor.Document == "" || invoice.Vendor.TaxDocument == "") {
		invoice.Vendor = *vendor
		invoice.VendorID = vendor.ID
	}

	if s.emitter == nil {
		s.logger.IPrintf(1, "Tax emitter is not configured or disabled")
		return dto.NewInvoiceTaxEmitResponse(503, "error", "Serviço de emissão fiscal (TaxEmitter) não está configurado", nil, nil, nil, 0)
	}

	// Calculate the next RPS number
	lastRps, err := s.repo.GetEmissionLastRPS(vendor.ID)
	if err != nil || lastRps == 0 {
		lastRps = vendor.LastRps
	}
	nextRPS := lastRps + 1

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	s.logger.IPrintf(1, "Calling TaxEmitter for invoice ID %d, RPS #%d", invoice.ID, nextRPS)
	emissionItem, err := s.emitter.EmitSingle(ctx, invoice, nextRPS)
	if err != nil {
		s.logger.IPrintf(1, "TaxEmitter error for invoice ID %d: %v", invoice.ID, err)
		return dto.NewInvoiceTaxEmitResponse(400, "error", fmt.Sprintf("Falha ao emitir NFS-e na Prefeitura: %v", err), nil, nil, nil, nextRPS)
	}

	// Persist the emission and update invoice & vendor in a database transaction
	if err := s.saveTaxEmission(vendor, invoice, emissionItem, nextRPS); err != nil {
		s.logger.IPrintf(1, "Failed to save tax emission for invoice ID %d: %v", invoice.ID, err)
		return dto.NewInvoiceTaxEmitResponse(500, "error", fmt.Sprintf("Nota emitida com sucesso na prefeitura (NFe #%d), mas ocorreu erro ao salvar no banco local: %v", *emissionItem.NFENumber, err), emissionItem.NFENumber, emissionItem.NFEVerification, nil, nextRPS)
	}

	nowStr := emissionItem.NFEDatetime.Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("NFS-e #%d emitida com sucesso (Cód. Verificação: %s)", *emissionItem.NFENumber, *emissionItem.NFEVerification)
	s.logger.IPrintf(1, "Invoice ID %d tax emitted successfully: %s", invoice.ID, msg)

	return dto.NewInvoiceTaxEmitResponse(200, "success", msg, emissionItem.NFENumber, emissionItem.NFEVerification, &nowStr, nextRPS)
}

func (s *InvoiceTaxEmit) saveTaxEmission(vendor *domain.Vendor, invoice *domain.Invoice, item *domain.EmissionItem, rpsNumber int64) error {
	if err := s.repo.BeginTransaction(); err != nil {
		return fmt.Errorf("falha ao iniciar transação no banco: %w", err)
	}
	defer s.repo.RollbackTransaction()

	now := time.Now()
	if item != nil && item.NFEDatetime != nil {
		now = *item.NFEDatetime
	}
	emission := domain.NewEmission(
		vendor,
		invoice.InvoiceDate,
		invoice.InvoiceDate,
		now,
		rpsNumber,
		rpsNumber,
		invoice.Amount,
		1,
		[]*domain.EmissionItem{item},
	)
	emission.NFEStart = item.NFENumber
	emission.NFEEnd = item.NFENumber
	emission.NFEDatetime = item.NFEDatetime

	if err := s.repo.Save(emission); err != nil {
		return fmt.Errorf("falha ao salvar registro de emissão: %w", err)
	}

	item.EmissionID = emission.ID
	if err := s.repo.Save(item); err != nil {
		return fmt.Errorf("falha ao salvar item da emissão: %w", err)
	}

	// Update invoice tax date
	invoice.TaxDate = item.NFEDatetime
	invoice.UpdatedAt = now
	if err := s.repo.Save(invoice); err != nil {
		return fmt.Errorf("falha ao atualizar tax_date na fatura: %w", err)
	}

	// Update vendor last RPS
	vendor.LastRps = rpsNumber
	vendor.UpdatedAt = now
	if err := s.repo.Save(vendor); err != nil {
		return fmt.Errorf("falha ao atualizar last_rps do fornecedor: %w", err)
	}

	if err := s.repo.CommitTransaction(); err != nil {
		return fmt.Errorf("falha ao efetivar transação de emissão: %w", err)
	}

	return nil
}
