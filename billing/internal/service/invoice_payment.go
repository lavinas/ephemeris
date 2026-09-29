package service

import (
	"fmt"
	"time"

	"billing/internal/dto"
	"billing/internal/port"
)

// InvoicePayment handles the business logic for registering an invoice payment and atomically sending the receipt email.
type InvoicePayment struct {
	*Base
	issuer port.Issuer
	pixer  port.Pixer
}

// NewInvoicePayment creates a new instance of InvoicePayment.
func NewInvoicePayment(repo port.Repository, logger port.Logger, issuer port.Issuer, pixer port.Pixer) *InvoicePayment {
	return &InvoicePayment{
		Base:   NewBase(repo, logger),
		issuer: issuer,
		pixer:  pixer,
	}
}

// Run executes the payment registration and receipt email dispatch in an atomic transaction.
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

	// Begin database transaction for atomicity
	if err := s.repo.BeginTransaction(); err != nil {
		s.logger.IPrintf(1, "Failed to begin transaction: %v", err)
		return dto.NewInvoicePaymentResponse(500, "error", "Failed to begin transaction: "+err.Error())
	}

	invoice := req.GetInvoice()
	paymentDate := req.GetPaymentDate()
	invoice.PaymentDate = &paymentDate
	invoice.UpdatedAt = time.Now()

	// Save payment date within transaction
	if err := s.repo.Save(invoice); err != nil {
		s.logger.IPrintf(1, "Failed to save invoice payment: %v", err)
		_ = s.repo.RollbackTransaction()
		return dto.NewInvoicePaymentResponse(500, "error", "Failed to record payment: "+err.Error())
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
	if billOut == nil || billOut.GetStatusCode() != 200 {
		errMsg := "Falha ao gerar recibo ou enviar e-mail"
		if billResp, ok := billOut.(*dto.BillResponse); ok && billResp.Message != "" {
			errMsg = billResp.Message
		}
		s.logger.IPrintf(1, "Receipt generation/email failed: %s. Rolling back payment transaction.", errMsg)
		_ = s.repo.RollbackTransaction()
		return dto.NewInvoicePaymentResponse(500, "error", errMsg)
	}

	// Commit transaction (persists payment date and email receipt date together)
	if err := s.repo.CommitTransaction(); err != nil {
		s.logger.IPrintf(1, "Failed to commit transaction: %v", err)
		_ = s.repo.RollbackTransaction()
		return dto.NewInvoicePaymentResponse(500, "error", "Failed to commit transaction: "+err.Error())
	}

	s.logger.IPrintf(2, "Successfully registered payment and sent receipt for invoice ID %d", invoice.ID)
	return dto.NewInvoicePaymentResponse(200, "success", "Payment registered and receipt sent successfully")
}
