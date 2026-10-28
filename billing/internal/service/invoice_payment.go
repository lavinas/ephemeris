package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"billing/internal/dto"
	"billing/internal/port"
)

const (
	// DefaultPaymentTimeout is the fallback timeout duration (20 seconds)
	DefaultPaymentTimeout = 20 * time.Second
)

var (
	globalTimeoutMutex   sync.RWMutex
	globalPaymentTimeout = DefaultPaymentTimeout
)

// SetGlobalPaymentTimeout sets the application-wide default timeout for invoice payment.
func SetGlobalPaymentTimeout(timeout time.Duration) {
	globalTimeoutMutex.Lock()
	defer globalTimeoutMutex.Unlock()
	if timeout > 0 {
		globalPaymentTimeout = timeout
	}
}

// GetGlobalPaymentTimeout retrieves the current application-wide default timeout.
func GetGlobalPaymentTimeout() time.Duration {
	globalTimeoutMutex.RLock()
	defer globalTimeoutMutex.RUnlock()
	return globalPaymentTimeout
}

// InvoicePayment handles the business logic for registering an invoice payment and atomically sending the receipt email.
type InvoicePayment struct {
	*Base
	issuer     port.Issuer
	pixer      port.Pixer
	taxEmitter port.TaxEmitter
	timeout    time.Duration
}

// NewInvoicePayment creates a new instance of InvoicePayment with the global configured timeout.
func NewInvoicePayment(repo port.Repository, logger port.Logger, issuer port.Issuer, pixer port.Pixer, taxEmitter ...port.TaxEmitter) *InvoicePayment {
	var emitter port.TaxEmitter
	if len(taxEmitter) > 0 {
		emitter = taxEmitter[0]
	}
	return &InvoicePayment{
		Base:       NewBase(repo, logger),
		issuer:     issuer,
		pixer:      pixer,
		taxEmitter: emitter,
		timeout:    GetGlobalPaymentTimeout(),
	}
}

// SetTaxEmitter configures the TaxEmitter instance for automatic NFS-e emission upon payment.
func (s *InvoicePayment) SetTaxEmitter(emitter port.TaxEmitter) *InvoicePayment {
	s.taxEmitter = emitter
	return s
}

// NewInvoicePaymentWithTimeout creates a new instance of InvoicePayment with a custom timeout.
func NewInvoicePaymentWithTimeout(repo port.Repository, logger port.Logger, issuer port.Issuer, pixer port.Pixer, timeout time.Duration) *InvoicePayment {
	if timeout <= 0 {
		timeout = GetGlobalPaymentTimeout()
	}
	return &InvoicePayment{
		Base:    NewBase(repo, logger),
		issuer:  issuer,
		pixer:   pixer,
		timeout: timeout,
	}
}

// SetTimeout sets a specific timeout duration for this InvoicePayment instance.
func (s *InvoicePayment) SetTimeout(timeout time.Duration) *InvoicePayment {
	if timeout > 0 {
		s.timeout = timeout
	}
	return s
}

// Run executes the payment registration and receipt email dispatch in an atomic transaction within the configured timeout.
func (s *InvoicePayment) Run(inDTO port.InDTO) port.OutDTO {
	req, ok := inDTO.(*dto.InvoicePaymentRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected InvoicePaymentRequest")
		return dto.NewInvoicePaymentResponse(400, "error", "Invalid input type")
	}
	s.logger.IPrintf(2, "Processing invoice payment request: %v", req)

	// Validate input against repository
	if err := req.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoicePaymentResponse(400, "error", fmt.Sprintf("Validation failed: %v", err))
	}

	timeout := s.timeout
	if timeout <= 0 {
		timeout = GetGlobalPaymentTimeout()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		resp port.OutDTO
	}
	resChan := make(chan result, 1)

	go func() {
		if ctx.Err() != nil {
			return
		}

		// Begin database transaction for atomicity
		if err := s.repo.BeginTransaction(); err != nil {
			s.logger.IPrintf(1, "Failed to begin transaction: %v", err)
			resChan <- result{resp: dto.NewInvoicePaymentResponse(500, "error", "Failed to begin transaction: "+err.Error())}
			return
		}

		if ctx.Err() != nil {
			_ = s.repo.RollbackTransaction()
			return
		}

		invoice := req.GetInvoice()
		paymentDate := req.GetPaymentDate()
		invoice.PaymentDate = &paymentDate
		invoice.UpdatedAt = time.Now()

		// Save payment date within transaction
		if err := s.repo.Save(invoice); err != nil {
			s.logger.IPrintf(1, "Failed to save invoice payment: %v", err)
			_ = s.repo.RollbackTransaction()
			resChan <- result{resp: dto.NewInvoicePaymentResponse(500, "error", "Failed to record payment: "+err.Error())}
			return
		}

		if ctx.Err() != nil {
			_ = s.repo.RollbackTransaction()
			return
		}

		// Generate and send receipt email via Bill service
		billService := NewBill(s.repo, s.logger, s.issuer, s.pixer)
		billReq := &dto.BillRequest{
			Vendor:    req.Vendor,
			InvoiceID: invoice.ID,
			Doc:       1, // 1 - Receipt
			Action:    0, // 0 - Send email
			Email:     "",
		}

		billOut := billService.Run(billReq)

		if ctx.Err() != nil {
			_ = s.repo.RollbackTransaction()
			return
		}

		if billOut == nil || billOut.GetStatusCode() != 200 {
			errMsg := "Falha ao gerar recibo ou enviar e-mail"
			if billResp, ok := billOut.(*dto.BillResponse); ok && billResp.Message != "" {
				errMsg = billResp.Message
			}
			s.logger.IPrintf(1, "Receipt generation/email failed: %s. Rolling back payment transaction.", errMsg)
			_ = s.repo.RollbackTransaction()
			resChan <- result{resp: dto.NewInvoicePaymentResponse(500, "error", errMsg)}
			return
		}

		// Commit transaction (persists payment date and email receipt date together)
		if err := s.repo.CommitTransaction(); err != nil {
			s.logger.IPrintf(1, "Failed to commit transaction: %v", err)
			_ = s.repo.RollbackTransaction()
			resChan <- result{resp: dto.NewInvoicePaymentResponse(500, "error", "Failed to commit transaction: "+err.Error())}
			return
		}

		successMsg := "Payment registered and receipt sent successfully"
		// If tax emitter is present and invoice is taxable, attempt automatic NFS-e emission
		if s.taxEmitter != nil && invoice.IsTaxable() {
			s.logger.IPrintf(2, "Invoice ID %d is taxable; triggering automatic NFS-e emission", invoice.ID)
			taxSvc := NewInvoiceTaxEmit(s.repo, s.logger, s.taxEmitter)
			taxReq := &dto.InvoiceTaxEmitRequest{
				Vendor: req.Vendor,
				ID:     invoice.ID,
			}
			taxOut := taxSvc.Run(taxReq)
			if taxOut != nil && taxOut.GetStatusCode() == 200 {
				if taxResp, ok := taxOut.(dto.InvoiceTaxEmitResponse); ok && taxResp.NFENumber != nil {
					successMsg = fmt.Sprintf("Pagamento registrado, recibo enviado e NFS-e #%d emitida com sucesso", *taxResp.NFENumber)
				} else {
					successMsg = "Pagamento registrado, recibo enviado e NFS-e emitida com sucesso"
				}
				s.logger.IPrintf(2, "Automatic NFS-e emission succeeded for invoice ID %d", invoice.ID)
			} else {
				errMsg := "Prefeitura indisponível ou rejeitou a emissão"
				if taxResp, ok := taxOut.(dto.InvoiceTaxEmitResponse); ok && taxResp.Message != "" {
					errMsg = taxResp.Message
				}
				s.logger.IPrintf(1, "Automatic NFS-e emission failed for invoice ID %d: %s. Invoice remains available for manual retry.", invoice.ID, errMsg)
				successMsg = fmt.Sprintf("Pagamento registrado e recibo enviado com sucesso. Aviso da NFS-e: %s", errMsg)
			}
		}

		s.logger.IPrintf(2, "Successfully registered payment and completed post-payment actions for invoice ID %d", invoice.ID)
		resChan <- result{resp: dto.NewInvoicePaymentResponse(200, "success", successMsg)}
	}()

	select {
	case res := <-resChan:
		return res.resp
	case <-ctx.Done():
		s.logger.IPrintf(1, "Invoice payment timed out after %v", timeout)
		_ = s.repo.RollbackTransaction()
		return dto.NewInvoicePaymentResponse(504, "error", "Tempo limite de processamento excedido (timeout). O pagamento não foi registrado.")
	}
}
