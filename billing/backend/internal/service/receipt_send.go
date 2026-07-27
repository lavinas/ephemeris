package service

import (
	"encoding/base64"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// ReceiptSend is responsible for handling the sending of receipts to customers.
type ReceiptSend struct {
	Base
	issuer port.Issuer
	piexer port.Pixer
}

// NewReceiptSend creates a new instance of ReceiptSend.
func NewReceiptSend(repo port.Repository, logger port.Logger, issuer port.Issuer, piexer port.Pixer) *ReceiptSend {
	return &ReceiptSend{
		Base:   *NewBase(repo, logger),
		issuer: issuer,
		piexer: piexer,
	}
}

// Run processes a request to send a receipt to a customer and returns the response.
func (s *ReceiptSend) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing send receipt request: %v", inDTO)
	// Type assertion to the expected DTO type
	in, ok := inDTO.(*dto.ReceiptSendRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected ReceiptSendRequest")
		return dto.NewReceiptSendResponse(400, "bad request", "Invalid input type", nil, nil)
	}
	// Validate the input DTO
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewReceiptSendResponse(400, "bad request", "Validation failed: "+err.Error(), nil, nil)
	}
	// Retrieve vendor and invoice from the repository
	vendor, err := s.repo.GetVendor(in.Vendor)
	if err != nil {
		return dto.NewReceiptSendResponse(400, "bad request", "vendor not found", nil, nil)
	}
	invoice, err := s.repo.GetInvoice(in.InvoiceID)
	if err != nil {
		return dto.NewReceiptSendResponse(400, "bad request", "invoice not found", nil, nil)
	}
	return s.dispatchAction(in, vendor, invoice)
}

// dispatchAction dispatches the action specified in the ReceiptSendRequest to the appropriate handler.
func (s *ReceiptSend) dispatchAction(in *dto.ReceiptSendRequest, vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	switch in.Action {
	case 0: // Send email
		return s.sendEmail(in, vendor, invoice)
	case 1: // Resend email
		return s.resendEmail(in, vendor, invoice)
	case 2: // Get PDF base64
		return s.getPDFBase64(in, vendor, invoice)
	default:
		return dto.NewReceiptSendResponse(400, "bad request", "invalid action", nil, nil)
	}
}

// sendEmail handles the action of sending a receipt email to the customer.
func (s *ReceiptSend) sendEmail(in *dto.ReceiptSendRequest, vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	issuerData := s.getIssuerData(vendor, invoice, in.Email)
	err := s.issuer.SendMail(issuerData, s.getHTMLTemplate("pdf"), s.getHTMLTemplate("email"))
	if err != nil {
		s.logger.IPrintf(2, "Error sending receipt: %v", err)
		return dto.NewReceiptSendResponse(500, "internal server error", "contact support", nil, nil)
	}
	s.registerSendReceiver(invoice)
	return dto.NewReceiptSendResponse(200, "success", "", nil, nil)
}

// resendEmail handles the action of resending a receipt email to the customer.
func (s *ReceiptSend) resendEmail(in *dto.ReceiptSendRequest, vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	issuerData := s.getIssuerData(vendor, invoice, in.Email)
	err := s.issuer.SendMail(issuerData, s.getHTMLTemplate("pdf"), s.getHTMLTemplate("email"))
	if err != nil {
		s.logger.IPrintf(2, "Error resending receipt: %v", err)
		return dto.NewReceiptSendResponse(500, "internal server error", "contact support", nil, nil)
	}
	return dto.NewReceiptSendResponse(200, "success", "", nil, nil)
}

// getPDFBase64 handles the action of generating a PDF receipt and returning it as a base64 string.
func (s *ReceiptSend) getPDFBase64(in *dto.ReceiptSendRequest, vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	issuerData := s.getIssuerData(vendor, invoice, in.Email)
	pdfBytes, err := s.issuer.GetBase64(issuerData, s.getHTMLTemplate("pdf"))
	if err != nil {
		s.logger.IPrintf(2, "Error generating PDF: %v", err)
		return dto.NewReceiptSendResponse(500, "internal server error", "contact support", nil, nil)
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)
	documentName := s.issuer.GetName(issuerData)
	return dto.NewReceiptSendResponse(200, "success", "", &pdfBase64, &documentName)
}

// registerSendReceiver registers the ReceiptSend service with the provided service registry.
func (s *ReceiptSend) registerSendReceiver(invoice *domain.Invoice) {
	invoice.UpdatedAt = time.Now()
	invoice.EmailReceiptDate = &invoice.UpdatedAt
	if err := s.repo.Save(invoice); err != nil {
		s.logger.IPrintf(2, "Error updating invoice after sending receipt: %v", err)
	}
}

// getIssuerData creates an IssuerData DTO based on the provided ReceiptSendRequest, Vendor, and Invoice.
func (s *ReceiptSend) getIssuerData(vendor *domain.Vendor, invoice *domain.Invoice, email string) *dto.IssuerData {
	items, totalAmount := s.getReceiptItems(invoice)
	sendEmail := invoice.Customer.Email
	if email != "" {
		sendEmail = &email
	}
	return &dto.IssuerData{
		VendorLogoBase64:   s.getLogoBase64(vendor),
		VendorName:         vendor.TradingName,
		VendorDocument:     vendor.Document,
		VendorEmail:        vendor.Email,
		VendorWhatsApp:     vendor.Whatsapp,
		VendorSMTPHost:     vendor.SmtpHost,
		VendorSMTPPort:     vendor.SmtpPort,
		VendorSMTPUsername: vendor.SmtpUser,
		VendorSMTPPassword: vendor.SmtpPassword,
		InvoiceNumber:      invoice.ID,
		CustomerName:       invoice.Customer.Name,
		CustomerNickname:   invoice.Customer.Nickname,
		CustomerDocument:   invoice.Customer.Document,
		CustomerEmail:      sendEmail,
		InvoiceDate:        invoice.InvoiceDate,
		DueDate:            invoice.DueDate,
		PaymentDate:        invoice.PaymentDate,
		Items:              items,
		TotalAmount:        totalAmount,
	}
}

// getLogoBase64 retrieves the base64-encoded logo for the given vendor.
func (s *ReceiptSend) getLogoBase64(vendor *domain.Vendor) template.URL {
	logo := filepath.Join(logoPath, vendor.LogoName)
	data, err := os.ReadFile(logo)
	if err != nil {
		s.logger.IPrintf(2, "Error reading logo file: %v", err)
		return ""
	}
	return template.URL(base64.StdEncoding.EncodeToString(data))
}

// getReceiptItems returns a slice of ReceiptItem DTOs for the given invoice.
func (s *ReceiptSend) getReceiptItems(invoice *domain.Invoice) ([]dto.ReceiptItem, float64) {
	var items []dto.ReceiptItem
	var totalAmount float64
	for _, item := range invoice.InvoiceItems {
		total := float64(item.Quantity) * item.Price
		items = append(items, dto.ReceiptItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.Price,
			Total:       total,
		})
		totalAmount += total
	}
	return items, totalAmount
}

// getHTMLTemplate retrieves the HTML template content for the given template type (pdf or email).
func (s *ReceiptSend) getHTMLTemplate(templateType string) string {
	path := filepath.Join(templatePath, "receipt_"+templateType+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		s.logger.IPrintf(2, "Error reading HTML template file: %v", err)
		return ""
	}
	return string(data)
}
