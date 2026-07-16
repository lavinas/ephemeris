package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// BillSendService is a service for sending bills to customers.
type BillGet struct {
	Base
	biller port.Biller
	pixer  port.Pixer
}

const (
	logoPath = "./images/"
)

// NewBillGet creates a new instance of BillGet.
func NewBillGet(repo port.Repository, logger port.Logger, biller port.Biller, pixer port.Pixer) *BillGet {
	return &BillGet{
		Base:   *NewBase(repo, logger),
		biller: biller,
		pixer:  pixer,
	}
}

// Get processes a request to send a bill to a customer and returns the response.
func (s *BillGet) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing send bill request: %v", inDTO)
	// Type assertion to the expected DTO type
	in, ok := inDTO.(*dto.BillGetRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected BillGetRequest")
		return dto.NewBillGetResponse(400, "bad request", "Invalid input type", 0, "", "")
	}
	s.logger.IPrintf(2, "Validated input: %v", in)
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewBillGetResponse(400, "bad request",
			"Validation failed: "+err.Error(), 0, "", "")
	}
	s.logger.IPrintf(2, "Input validated successfully: %v", in)
	document, name, err := s.getDocument(in.DocumentType, in.Vendor, in.InvoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Error generating document: %v", err)
		return dto.NewBillGetResponse(500, "internal server error",
			"contact support", 0, "", "")
	}
	s.logger.IPrintf(2, "Document generated successfully: %v", document)
	return dto.NewBillGetResponse(200, "success", "", in.DocumentType, document, name)
}

// getDocumentName returns the document name based on the document type and invoice ID.
func (s *BillGet) getDocumentName(docType int, invoice *domain.Invoice) string {
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

// getDocument is a helper function to retrieve the document based on the document type.
func (s *BillGet) getDocument(docType int, vendorNick string, invoiceID int64) (string, string, error) {
	s.logger.IPrintf(2, "Retrieving document for type: %d, vendor: %s, invoiceID: %d", docType, vendorNick, invoiceID)
	vendor, err := s.repo.GetVendor(vendorNick)
	if err != nil {
		return "", "", errors.New("vendor not found")
	}
	invoice, err := s.repo.GetInvoice(invoiceID)
	if err != nil {
		return "", "", errors.New("invoice not found")
	}
	var document string
	switch docType {
	case 1:
		document, err = s.getBill(vendor, invoice)
	case 2:
		document, err = s.getQRCode(vendor, invoice)
	case 3:
		document, err = s.getPayload(vendor, invoice)
	default:
		return "", "", errors.New("invalid document type")
	}
	if err != nil {
		return "", "", err
	}
	s.logger.IPrintf(2, "Document retrieved successfully for type: %d, vendor: %s, invoiceID: %d", docType, vendorNick, invoiceID)
	return document, s.getDocumentName(docType, invoice), nil
}

// getPixStrings is a helper function to retrieve the Pix strings based on the document type.
func (s *BillGet) getPixStrings(invoice *domain.Invoice, vendor *domain.Vendor) (*string, *string, error) {
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
	s.logger.IPrintf(2, "PixRequest DTO created: %v", dto)
	payload, qrCode, err := s.pixer.Get(dto)
	s.logger.IPrintf(2, "Pix strings generated: payload=%v, qrCode=%v, error=%v", payload, qrCode, err)
	if err != nil {
		return nil, nil, err
	}
	s.logger.IPrintf(2, "Pix strings generated successfully: payload=%v, qrCode=%v", payload, qrCode)
	return &payload, &qrCode, nil
}

// getBill generates the bill document for the given invoice.
func (s *BillGet) getBill(vendor *domain.Vendor, invoice *domain.Invoice) (string, error) {
	s.logger.IPrintf(2, "Generating bill for vendor: %v, invoice: %v", vendor, invoice)
	payload, qrCode, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		return "", err
	}
	if payload == nil || qrCode == nil {
		return "", errors.New("vendor does not have a Pix token")
	}
	billDto := s.getBillDto(invoice, vendor, *payload, *qrCode)
	s.logger.IPrintf(2, "Bill DTO created: %v", billDto)
	bill, err := s.biller.GetPDFBase64(billDto)
	if err != nil {
		return "", err
	}
	s.logger.IPrintf(2, "Generating bill document using Maroto for invoice: %v", invoice)
	return bill, nil
}

// getQRCode generates the QR code document for the given invoice.
func (s *BillGet) getQRCode(vendor *domain.Vendor, invoice *domain.Invoice) (string, error) {
	s.logger.IPrintf(2, "Generating QR code for vendor: %v, invoice: %v", vendor, invoice)
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
func (s *BillGet) getPayload(vendor *domain.Vendor, invoice *domain.Invoice) (string, error) {
	payload, _, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		return "", err
	}
	if payload == nil {
		return "", errors.New("vendor does not have a Pix token")
	}
	return *payload, nil
}

// getBillDto creates a BillRequest DTO for the given invoice and vendor.
func (s *BillGet) getBillDto(invoice *domain.Invoice, vendor *domain.Vendor, payload string, qrCode string) *dto.BillerRequest {
	s.logger.IPrintf(2, "Creating BillRequest DTO for vendor: %v, invoice: %v", vendor, invoice)
	logo := filepath.Join(logoPath, vendor.LogoName)

	return &dto.BillerRequest{
		InvoiceID:   invoice.ID,
		InvoiceDate: invoice.InvoiceDate,
		InvoiceDue:  invoice.DueDate,
		Vendor: dto.BillerVendor{
			Logo:     logo,
			Name:     vendor.LegalName,
			Document: vendor.Document,
			Email:    &vendor.Email,
			Whatsapp: &vendor.Whatsapp,
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
	}
}

// getBillItems creates a slice of BillerItem DTOs for the given invoice.
func (s *BillGet) getBillItems(invoice *domain.Invoice) []dto.BillerItem {
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
