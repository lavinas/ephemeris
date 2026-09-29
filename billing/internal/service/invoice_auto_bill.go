package service

import (
	"fmt"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// InvoiceAutoBill is responsible for finding invoices due within N days and sending them by email.
type InvoiceAutoBill struct {
	Base
	billService   *Bill
	daysInAdvance int
}

// NewInvoiceAutoBill creates a new instance of InvoiceAutoBill.
func NewInvoiceAutoBill(repo port.Repository, logger port.Logger, billService *Bill, daysInAdvance int) *InvoiceAutoBill {
	if daysInAdvance <= 0 {
		daysInAdvance = 3
	}
	return &InvoiceAutoBill{
		Base:          *NewBase(repo, logger),
		billService:   billService,
		daysInAdvance: daysInAdvance,
	}
}

// Run executes the auto billing process.
func (s *InvoiceAutoBill) Run(inDTO port.InDTO) port.OutDTO {
	days := s.daysInAdvance
	now := time.Now()

	if inDTO != nil {
		if req, ok := inDTO.(*dto.InvoiceAutoBillRequest); ok && req != nil {
			if req.DaysInAdvance > 0 {
				days = req.DaysInAdvance
			}
			if req.ReferenceDate != nil {
				now = *req.ReferenceDate
			}
		}
	}

	// Calculate target due date limit: end of the day (now + days)
	dueLimit := time.Date(now.Year(), now.Month(), now.Day()+days, 23, 59, 59, 999999999, now.Location())
	s.logger.IPrintf(1, "InvoiceAutoBill: Searching pending invoices with DueDate <= %s (advance days: %d)",
		dueLimit.Format("2006-01-02 15:04:05"), days)

	invoices, err := s.repo.FindInvoicesPendingSend(dueLimit)
	if err != nil {
		s.logger.IPrintf(0, "InvoiceAutoBill: Error searching pending invoices: %v", err)
		return dto.NewInvoiceAutoBillResponse(500, "internal error", err.Error(), 0, 0, 0, nil)
	}

	totalFound := len(invoices)
	s.logger.IPrintf(1, "InvoiceAutoBill: Found %d pending invoice(s) to process", totalFound)

	if totalFound == 0 {
		return dto.NewInvoiceAutoBillResponse(200, "success", "No pending invoices to send", 0, 0, 0, nil)
	}

	totalSent := 0
	totalErrors := 0
	var errorDetails []string

	// Cache vendor lookups by ID
	vendorMap := make(map[int64]*domain.Vendor)

	for _, inv := range invoices {
		vendor, ok := vendorMap[inv.Customer.VendorID]
		if !ok {
			v, err := s.repo.GetVendorByID(inv.Customer.VendorID)
			if err != nil || v == nil {
				errMsg := fmt.Sprintf("Invoice %d: vendor ID %d not found", inv.ID, inv.Customer.VendorID)
				if err != nil {
					errMsg = fmt.Sprintf("Invoice %d: error fetching vendor ID %d: %v", inv.ID, inv.Customer.VendorID, err)
				}
				s.logger.IPrintf(0, "InvoiceAutoBill: %s", errMsg)
				errorDetails = append(errorDetails, errMsg)
				totalErrors++
				continue
			}
			vendor = v
			vendorMap[inv.Customer.VendorID] = vendor
		}

		// Prepare BillRequest (Action 0 = send email, Doc 0 = Invoice)
		billReq := &dto.BillRequest{
			Vendor:    vendor.Nickname,
			InvoiceID: inv.ID,
			Doc:       0,
			Action:    0,
		}

		outDTO := s.billService.Run(billReq)
		if outDTO.GetStatusCode() != 200 {
			errMsg := fmt.Sprintf("Invoice %d: failed to send email (status %d)", inv.ID, outDTO.GetStatusCode())
			if billResp, ok := outDTO.(*dto.BillResponse); ok {
				errMsg = fmt.Sprintf("Invoice %d: failed to send email: %s (%s)", inv.ID, billResp.Status, billResp.Message)
			}
			s.logger.IPrintf(0, "InvoiceAutoBill: %s", errMsg)
			errorDetails = append(errorDetails, errMsg)
			totalErrors++
			continue
		}

		s.logger.IPrintf(1, "InvoiceAutoBill: Invoice %d successfully sent to customer %s (%s)",
			inv.ID, inv.Customer.Nickname, vendor.Nickname)
		totalSent++
	}

	s.logger.IPrintf(1, "InvoiceAutoBill: Process completed. Total: %d, Sent: %d, Errors: %d",
		totalFound, totalSent, totalErrors)

	return dto.NewInvoiceAutoBillResponse(200, "success", "Auto bill process finished", totalFound, totalSent, totalErrors, errorDetails)
}
