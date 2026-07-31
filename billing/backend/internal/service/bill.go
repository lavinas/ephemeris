package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
)

// Bill is responsible for handling the sending of receipts to customers.
type Bill struct {
	Base
	issuer port.Issuer
	pixer  port.Pixer
}

// NewBill creates a new instance of Bill.
func NewBill(repo port.Repository, logger port.Logger,
	issuer port.Issuer, pixer port.Pixer) *Bill {
	return &Bill{
		Base:   *NewBase(repo, logger),
		issuer: issuer,
		pixer:  pixer,
	}
}

// Run processes a request to send a receipt to a customer and returns the response.
func (s *Bill) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing send receipt request: %v", inDTO)
	// Type assertion to the expected DTO type
	in, ok := inDTO.(*dto.BillRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected BillRequest")
		return dto.NewBillResponse(400, "bad request", "Invalid input type", nil, nil)
	}
	// Validate the input DTO
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewBillResponse(400, "bad request", "Validation failed: "+err.Error(), nil, nil)
	}
	// Retrieve vendor and invoice from the repository
	vendor, err := s.repo.GetVendor(in.Vendor)
	if err != nil {
		return dto.NewBillResponse(400, "bad request", "vendor not found", nil, nil)
	}
	invoice, err := s.repo.GetInvoice(in.InvoiceID)
	if err != nil {
		return dto.NewBillResponse(400, "bad request", "invoice not found", nil, nil)
	}
	return s.dispatchAction(in, vendor, invoice)
}

// dispatchAction dispatches the action specified in the BillRequest to the appropriate handler.
func (s *Bill) dispatchAction(in *dto.BillRequest,
	vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {

	switch in.Action {
	case 0: // Send email
		return s.sendEmail(in, vendor, invoice)
	case 1: // Resend email
		return s.resendEmail(in, vendor, invoice)
	case 2: // Get PDF base64
		return s.getPDFBase64(in, vendor, invoice)
	default:
		return dto.NewBillResponse(400, "bad request", "invalid action", nil, nil)
	}
}

// getsubject returns the email subject based on the action specified in the BillRequest.
func (s *Bill) getSubject(docType int, invoiceDate time.Time) string {
	subMap := map[int]string{
		0: "Estúdio Amelia Cardoso - sua fatura de %s de %s",
		1: "Estúdio Amelia Cardoso - recibo de sua fatura de %s de %s",
	}
	monthsMap := map[int]string{
		1:  "Janeiro",
		2:  "Fevereiro",
		3:  "Março",
		4:  "Abril",
		5:  "Maio",
		6:  "Junho",
		7:  "Julho",
		8:  "Agosto",
		9:  "Setembro",
		10: "Outubro",
		11: "Novembro",
		12: "Dezembro",
	}
	ret, ok := subMap[docType]
	if !ok {
		return ""
	}
	month := monthsMap[int(invoiceDate.Month())]
	return fmt.Sprintf(ret, month, invoiceDate.Format("2006"))
}

// sendEmail handles the action of sending a receipt email to the customer.
func (s *Bill) sendEmail(in *dto.BillRequest,
	vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	dto := s.resendEmail(in, vendor, invoice)
	s.registerSendReceiver(invoice, in.Doc)
	return dto
}

// resendEmail handles the action of resending a receipt email to the customer.
func (s *Bill) resendEmail(in *dto.BillRequest,
	vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	subject := s.getSubject(in.Doc, invoice.InvoiceDate)
	issuerData, err := s.getIssuerData(vendor, invoice, in.Email)
	if err != nil {
		s.logger.IPrintf(2, "Error getting issuer data: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	pdf_template, err := s.getHTMLTemplate("pdf", in.Doc)
	if err != nil {
		s.logger.IPrintf(2, "Error getting PDF template: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	email_template, err := s.getHTMLTemplate("email", in.Doc)
	if err != nil {
		s.logger.IPrintf(2, "Error getting email template: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	err = s.issuer.SendMail(issuerData, subject, pdf_template, email_template)
	if err != nil {
		s.logger.IPrintf(2, "Error resending receipt: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	return dto.NewBillResponse(200, "success", "", nil, nil)
}

// getPDFBase64 handles the action of generating a PDF receipt and returning it as a base64 string.
func (s *Bill) getPDFBase64(in *dto.BillRequest,
	vendor *domain.Vendor, invoice *domain.Invoice) port.OutDTO {
	issuerData, err := s.getIssuerData(vendor, invoice, in.Email)
	if err != nil {
		s.logger.IPrintf(2, "Error getting issuer data: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	pdf_template, err := s.getHTMLTemplate("pdf", in.Doc)
	if err != nil {
		s.logger.IPrintf(2, "Error getting PDF template: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	pdfBytes, err := s.issuer.GetBase64(issuerData, pdf_template)
	if err != nil {
		s.logger.IPrintf(2, "Error generating PDF: %v", err)
		return dto.NewBillResponse(500, "internal server error", "contact support", nil, nil)
	}
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)
	documentName := s.issuer.GetName(issuerData)
	return dto.NewBillResponse(200, "success", "", &pdfBase64, &documentName)
}

// registerSendReceiver registers the Bill service with the provided service registry.
func (s *Bill) registerSendReceiver(invoice *domain.Invoice, docType int) {
	invoice.UpdatedAt = time.Now()
	switch docType {
	case 0: // Invoice
		invoice.EmailSentDate = &invoice.UpdatedAt
	case 1: // Receipt
		invoice.EmailReceiptDate = &invoice.UpdatedAt
	}
	if err := s.repo.Save(invoice); err != nil {
		s.logger.IPrintf(2, "Error updating invoice after sending receipt: %v", err)
	}
}

// getIssuerData creates an IssuerData DTO based on the provided BillRequest, Vendor, and Invoice.
func (s *Bill) getIssuerData(vendor *domain.Vendor,
	invoice *domain.Invoice, email string) (*dto.IssuerData, error) {
	items, totalAmount := s.getReceiptItems(invoice)
	sendEmail := ""
	if invoice.Customer.Email != nil {
		sendEmail = *invoice.Customer.Email
	}
	if email != "" {
		sendEmail = email
	}
	document := ""
	if invoice.Customer.Document != nil {
		document = *invoice.Customer.Document
	}
	payload, qrcode, err := s.getPixStrings(invoice, vendor)
	if err != nil {
		s.logger.IPrintf(2, "Error generating Pix strings: %v", err)
		return nil, err
	}
	return &dto.IssuerData{
		VendorLogoBase64:     s.getLogoBase64(vendor),
		VendorName:           vendor.TradingName,
		VendorDocument:       vendor.Document,
		VendorEmail:          vendor.Email,
		VendorWhatsApp:       vendor.Whatsapp,
		VendorSMTPHost:       vendor.SmtpHost,
		VendorSMTPPort:       vendor.SmtpPort,
		VendorSMTPUsername:   vendor.SmtpUser,
		VendorSMTPPassword:   vendor.SmtpPassword,
		VendorPixQRBase64:    template.URL(qrcode),
		VendorPixCopyPaste:   payload,
		VendorPixName:        vendor.PixName,
		VendorBank:           vendor.AccountBank,
		VendorAgency:         vendor.AccountAgency,
		VendorAccount:        vendor.AccountNumber,
		InvoiceNumber:        invoice.ID,
		CustomerFirstName:    s.getFirstName(invoice.Customer.Name),
		CustomerName:         invoice.Customer.Name,
		CustomerNickname:     invoice.Customer.Nickname,
		CustomerDocumentType: s.getDocumentType(invoice.Customer.Document),
		CustomerDocument:     document,
		CustomerEmail:        sendEmail,
		InvoiceDate:          invoice.InvoiceDate,
		InvoiceDueDate:       invoice.DueDate,
		InvoicePaymentDate:   s.getPaymentDate(invoice.PaymentDate),
		InvoiceItems:         items,
		InvoiceTotalAmount:   totalAmount,
	}, nil
}

// getPixStrings generates the Pix payload and QR code strings for the given invoice and vendor.
func (s *Bill) getPixStrings(invoice *domain.Invoice, vendor *domain.Vendor) (string, string, error) {
	s.logger.IPrintf(2, "Generating Pix strings for vendor: %v, invoice: %v", vendor, invoice)
	if vendor.PixToken == "" {
		return "", "", nil
	}
	if invoice == nil || len(invoice.InvoiceItems) == 0 {
		return "", "", errors.New("invoice is invalid or has no items")
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
		return "", "", err
	}
	s.logger.IPrintf(2, "Pix strings generated successfully")
	return payload, qrCode, nil
}

// getFirstName returns the first name from a full name string.
func (s *Bill) getFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// getPaymentDate returns the payment date if available, otherwise returns nil.
func (s *Bill) getPaymentDate(paymentDate *time.Time) time.Time {
	if paymentDate != nil {
		return *paymentDate
	}
	return time.Time{}
}

// getDocumentType returns the type of document based on its length.
func (s *Bill) getDocumentType(document *string) string {
	if document == nil {
		return ""
	}
	if len(*document) >= 14 {
		return "CNPJ"
	}
	return "CPF"
}

// getLogoBase64 retrieves the base64-encoded logo for the given vendor.
func (s *Bill) getLogoBase64(vendor *domain.Vendor) template.URL {
	logo := filepath.Join(logoPath, vendor.LogoName)
	data, err := os.ReadFile(logo)
	if err != nil {
		s.logger.IPrintf(2, "Error reading logo file: %v", err)
		return ""
	}
	return template.URL(base64.StdEncoding.EncodeToString(data))
}

// getReceiptItems returns a slice of ReceiptItem DTOs for the given invoice.
func (s *Bill) getReceiptItems(invoice *domain.Invoice) ([]dto.ReceiptItem, float64) {
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
func (s *Bill) getHTMLTemplate(templateType string, doc int) (string, error) {
	docType := map[int]string{
		0: "invoice",
		1: "receipt",
	}
	docName, ok := docType[doc]
	if !ok {
		return "", fmt.Errorf("invalid document type: %d", doc)
	}
	path := filepath.Join(templatePath, docName+"_"+templateType+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		s.logger.IPrintf(2, "Error reading HTML template file: %v", err)
		return "", err
	}
	return string(data), nil
}
