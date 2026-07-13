package service

import (
	"errors"
	"strconv"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/domain"
)

// BillSendService is a service for sending bills to customers.
type InvoiceBillService struct {
	Base
	biller port.Biller
	pixer  port.Pixer
}

// NewInvoiceBillService creates a new instance of InvoiceBillService.
func NewInvoiceBillService(repo port.Repository, logger port.Logger, biller port.Biller, pixer port.Pixer) *InvoiceBillService {
	return &InvoiceBillService{
		Base:   *NewBase(repo, logger),
		biller: biller,
		pixer:  pixer,
	}
}

// Get processes a request to send a bill to a customer and returns the response.
func (s *InvoiceBillService) Get(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing send bill request: %v", inDTO)
	// Type assertion to the expected DTO type
	in, ok := inDTO.(*dto.GetBillRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected GetBillRequest")
		return dto.NewGetBillResponse(400, "bad request", "Invalid input type", 0, "")
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewGetBillResponse(400, "bad request",
			"Validation failed: "+err.Error(), 0, "")
	}
	documentBase64, err := s.getDocument(in.DocumentType, in.Vendor, in.InvoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Error generating document: %v", err)
		return dto.NewGetBillResponse(500, "internal server error",
			"contact support", 0, "")
	}
	return dto.NewGetBillResponse(200, "success", "", in.DocumentType, documentBase64)
}

// getDocument is a helper function to retrieve the document based on the document type.
func (s *InvoiceBillService) getDocument(docType int, vendorNick string, invoiceID int64) (string, error) {
	vendor, err := s.repo.GetVendor(vendorNick)
	if err != nil {
		return "", errors.New("vendor not found")
	}
	invoice , err := s.repo.GetInvoice(invoiceID)
	if err != nil {
		return "", errors.New("invoice not found")
	}
	switch docType {
	case 1:
		return s.getBill(vendor, invoice)
	case 2:
		return s.getQRCode(vendor, invoice)
	case 3:
		return s.getPayload(vendor, invoice)
	default:
		return "", errors.New("invalid document type")
	}
}

// getPixStrings is a helper function to retrieve the Pix strings based on the document type.
func (s *InvoiceBillService) getPixStrings(invoice *domain.Invoice, vendor *domain.Vendor) (*string, *string, error) {
	if vendor.PixToken == "" {
		return nil, nil, nil
	}
	idStr := strconv.FormatInt(invoice.ID, 10)
	dto := &dto.PixRequest{
		Key:         vendor.PixToken,
		Description: "Payment for invoice",
		Name:        vendor.LegalName,
		City:        vendor.LegalName,
		Txid:        idStr, // Use invoice ID as Txid
		Amount:      invoice.Amount, // Example amount, replace with actual value
	}
	payload, qrCode, err := s.pixer.Get(dto)
	if err != nil {
		return nil, nil, err
	}
	return &payload, &qrCode, nil
}


// getBill generates the bill document for the given invoice.
func (s *InvoiceBillService) getBill(vendor *domain.Vendor, invoice *domain.Invoice) (string, error) {
	payload, _, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		return "", err
	}
	if payload == nil {
		return "", errors.New("vendor does not have a Pix token")
	}	
	return *payload, nil
}

// getQRCode generates the QR code document for the given invoice.
func (s *InvoiceBillService) getQRCode(vendor *domain.Vendor, invoice *domain.Invoice) (string, error) {
	_, qrCode, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		return "", err
	}
	if qrCode == nil {
		return "", errors.New("vendor does not have a Pix token")
	}	
	return *qrCode, nil
}

// getPayload generates the Pix document for the given invoice.
func (s *InvoiceBillService) getPayload(vendor *domain.Vendor, invoice *domain.Invoice) (string, error) {
	payload, _, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		return "", err
	}
	if payload == nil {
		return "", errors.New("vendor does not have a Pix token")
	}	
	return *payload, nil
}