package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// BillSend is responsible for handling the sending of bills to customers.
type BillSend struct {
	Base
	biller port.Biller
	pixer  port.Pixer
}

// NewBillSend creates a new instance of BillSend.
func NewBillSend(repo port.Repository, logger port.Logger, biller port.Biller, pixer port.Pixer) *BillSend {
	return &BillSend{
		Base:   *NewBase(repo, logger),
		biller: biller,
		pixer:  pixer,
	}
}

// Run processes a request to send a bill to a customer and returns the response.
func (s *BillSend) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing send bill request: %v", inDTO)
	// Type assertion to the expected DTO type
	in, ok := inDTO.(*dto.BillSendRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected BillSendRequest")
		return dto.NewBillSendResponse(400, "bad request", "Invalid input type")
	}
	// Validate the input DTO
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewBillSendResponse(400, "bad request", "Validation failed: "+err.Error())
	}
	// Retrieve vendor and invoice from the repository
	vendor, err := s.repo.GetVendor(in.Vendor)
	if err != nil {
		return dto.NewBillSendResponse(400, "bad request", "vendor not found")
	}
	// Retrieve the invoice from the repository
	invoice, err := s.repo.GetInvoice(in.InvoiceID)
	if err != nil {
		return dto.NewBillSendResponse(400, "bad request", "invoice not found")
	}
	// Generate Pix strings for the invoice and vendor
	payload, qrCode, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		return dto.NewBillSendResponse(500, "internal server error", "contact support")
	}
	// Create the BillRequest DTO and send the bill via email
	sDto := s.getBillDto(invoice, vendor, *payload, *qrCode, in.SendCopy)
	if err := s.biller.SendMail(sDto); err != nil {
		s.logger.IPrintf(2, "Error sending bill: %v", err)
		return dto.NewBillSendResponse(500, "internal server error", "contact support")
	}
	// Update the invoice's EmailSentDate and save it to the repository
	if err := s.updateInvoiceEmailSentDate(invoice); err != nil {
		return dto.NewBillSendResponse(500, "internal server error", "contact support")
	}
	// Log the successful bill sending and return a success response
	s.logger.IPrintf(2, "Bill sent successfully for invoice ID: %v", in.InvoiceID)
	return dto.NewBillSendResponse(200, "success", "")
}

// / getBillDto creates a BillRequest DTO for the given invoice and vendor.
func (s *BillSend) getBillDto(invoice *domain.Invoice, vendor *domain.Vendor, payload string, qrCode string, sendCopy bool) *dto.BillerRequest {
	s.logger.IPrintf(2, "Creating BillRequest DTO for vendor: %v, invoice: %v", vendor, invoice)
	logo := filepath.Join(logoPath, vendor.LogoName)
	return &dto.BillerRequest{
		InvoiceID:    invoice.ID,
		InvoiceDate:  invoice.InvoiceDate,
		InvoiceDue:   invoice.DueDate,
		BillFileName: s.getDocumentName(1, invoice),
		SendCopy:     sendCopy,
		Vendor: dto.BillerVendor{
			Logo:        logo,
			LegalName:   vendor.LegalName,
			TradingName: vendor.TradingName,
			Document:    vendor.Document,
			Email:       &vendor.Email,
			Whatsapp:    &vendor.Whatsapp,
		},
		Customer: dto.BillerCustomer{
			Name:     invoice.Customer.Name,
			Document: invoice.Customer.Document,
			Email:    invoice.Customer.Email,
			Whatsapp: invoice.Customer.Whatsapp,
		},
		Items: s.getBillItems(invoice),
		Receive: dto.BillerReceive{
			Pix: &dto.BillerPix{
				PixKey:       vendor.PixToken,
				ReceiverName: vendor.PixName,
				PixCopyPaste: payload,
				PixQRCode:    qrCode,
			},
			BankAccount: &dto.BillerBankAccount{
				BankName:         vendor.AccountBank,
				BankAgency:       vendor.AccountAgency,
				BankAccount:      vendor.AccountNumber,
				ReceiverName:     vendor.LegalName,
				ReceiverDocument: vendor.Document,
			},
		},
		SMTP: dto.BillerSMTP{
			SmtpHost:     vendor.SmtpHost,
			SmtpPort:     vendor.SmtpPort,
			SmtpUser:     vendor.SmtpUser,
			SmtpPassword: vendor.SmtpPassword,
		},
	}
}

// getBillItems creates a slice of BillerItem DTOs for the given invoice.
func (s *BillSend) getBillItems(invoice *domain.Invoice) []dto.BillerItem {
	s.logger.IPrintf(2, "Creating BillerItem DTOs for invoice: %v", invoice)
	items := make([]dto.BillerItem, len(invoice.InvoiceItems))
	for i, item := range invoice.InvoiceItems {
		items[i] = dto.BillerItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Price:       item.Price,
		}
	}
	s.logger.IPrintf(2, "BillerItem DTOs created: %v", items)
	return items
}

// getDocumentName returns the document name based on the document type and invoice ID.
func (s *BillSend) getDocumentName(docType int, invoice *domain.Invoice) string {
	nameFmt := "%s-%s-%s.%s"
	extMap := map[int]string{
		1: "pdf",
		2: "png",
		3: "txt",
	}
	ext, _ := extMap[docType]
	invDate := invoice.InvoiceDate.Format("2006-01-02")
	dueDate := invoice.DueDate.Format("2006-01-02")
	return fmt.Sprintf(nameFmt, dueDate, invDate, invoice.Customer.Nickname, ext)
}

// getPixStrings is a helper function to retrieve the Pix strings based on the document type.
func (s *BillSend) getPixStrings(invoice *domain.Invoice, vendor *domain.Vendor) (*string, *string, error) {
	s.logger.IPrintf(2, "Generating Pix strings for vendor: %v, invoice: %v", vendor, invoice)
	if vendor.PixToken == "" {
		return nil, nil, nil
	}
	if invoice == nil || len(invoice.InvoiceItems) == 0 {
		return nil, nil, errors.New("invoice is invalid or has no items")
	}
	idStr := strconv.FormatInt(invoice.ID, 10)
	dto := &dto.PixRequest{
		Key:         vendor.PixToken,
		Description: invoice.InvoiceItems[0].Description,
		Name:        vendor.PixName,
		City:        vendor.PixCity,
		Txid:        idStr,          // Use invoice ID as Txid
		Amount:      invoice.Amount, // Example amount, replace with actual value
	}
	payload, qrCode, err := s.pixer.Get(dto)
	if err != nil {
		return nil, nil, err
	}
	s.logger.IPrintf(2, "Pix strings generated successfully")
	return &payload, &qrCode, nil
}

// updateInvoiceEmailSentDate updates the EmailSentDate of the invoice and saves it to the repository.
func (s *BillSend) updateInvoiceEmailSentDate(invoice *domain.Invoice) error {
	s.logger.IPrintf(2, "Updating EmailSentDate for invoice ID: %v", invoice.ID)
	now := time.Now()
	invoice.EmailSentDate = &now
	if err := s.repo.Save(invoice); err != nil {
		s.logger.IPrintf(2, "Error saving invoice: %v", err)
		return err
	}
	s.logger.IPrintf(2, "EmailSentDate updated successfully for invoice ID: %v", invoice.ID)
	return nil
}
