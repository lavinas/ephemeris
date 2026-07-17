package driven

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"
	"time"

	"billing/internal/dto"
	"billing/internal/port"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"gopkg.in/gomail.v2"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	mline "github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

const (
	bodyTemplate = `
	<html>
		<body style="font-family: Arial, sans-serif; font-size: 16px; color: #333;">
			<p>Olá {{.Name}},</p>
			<p>Obrigado pela confiança, sempre buscamos investir em aperfeiçoamento pedagógico e em tecnologia.</p>
			<p>Enviamos em anexo a fatura referente ao vencimento em <b>{{.DueDate}}</b> no valor de <b>{{.Amount}}</b>, referente a:.</p>
			{{.Description}}
			<p>Para sua comodidade, disponibilizamos também aqui o pagamento via Pix. Utilize 
				o QRCode ou copie e cole o código pix que estão abaixo:</p>
			<p><img src="cid:{{.CidQRCode}}" alt="QR Code" style="width: 150px; height: 150px;"></p>
			<p><span style="font-family: monospace; font-size: 14px; background-color: #f4f4f4; 
				padding: 10px; border-radius: 5px;">{{.CopyPasteCode}}</span></p>
			<p style="font-size: 14px">
			    <b>*</b>Por favor, antes de confirmar o pagamento, verifique que o valor é 
					{{.Amount}} e que o recebedor é {{.VendorName}}.
			</p>
			<p> </p>
			<p> </p>
			<p>Atenciosamente,</p>
			<p>{{.VendorName}}</p>
		</body>
	</html>
	`
)

// emailData represents the data used to populate the email template.
type bodyData struct {
	VendorName    string
	Name          string
	InvoiceDate   string
	DueDate       string
	Amount        string
	Description   template.HTML
	QrCodeBase64  template.HTML
	CopyPasteCode string
	CidQRCode     string
}

// PDFGenerator is an implementation of the PDFGenerator interface that generates PDF files.
type BillerMaroto struct {
	logger    port.Logger
	generator core.Maroto
}

// NewBillerMaroto creates a new instance of BillerMaroto.
func NewBillerMaroto(logger port.Logger) *BillerMaroto {
	biller := &BillerMaroto{
		logger: logger,
	}
	biller.resetGenerator()
	return biller
}

// GeneratePDF generates a PDF file based on the provided data and returns the file path.
func (p *BillerMaroto) Generate(request port.InDTO, path string) error {
	requestDTO, ok := request.(*dto.BillerRequest)
	if !ok {
		return fmt.Errorf("invalid request type: expected BillerRequest")
	}
	document, err := p.getForm(*requestDTO)
	if err != nil {
		return err
	}
	err = document.Save(path)
	if err != nil {
		return err
	}
	return nil
}

// GetPDFPath returns the path where the generated PDF will be saved.
func (p *BillerMaroto) GetBinary(request port.InDTO) ([]byte, error) {
	requestDTO, ok := request.(*dto.BillerRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected BillerRequest")
	}
	document, err := p.getForm(*requestDTO)
	if err != nil {
		return nil, err
	}
	return document.GetBytes(), nil
}

// GetPDFBase64 returns the base64 encoded string of the generated PDF.
func (p *BillerMaroto) GetPDFBase64(request port.InDTO) (string, error) {
	requestDTO, ok := request.(*dto.BillerRequest)
	if !ok {
		return "", fmt.Errorf("invalid request type: expected BillerRequest")
	}
	document, err := p.GetBinary(requestDTO)
	if err != nil {
		return "", err
	}
	doc := base64.StdEncoding.EncodeToString(document)
	return doc, nil
}

// SendMail generates a PDF file based on the provided data and sends it via email.
func (p *BillerMaroto) SendMail(request port.InDTO) error {
	requestDTO, ok := request.(*dto.BillerRequest)
	if !ok {
		return fmt.Errorf("invalid request type: expected BillerRequest")
	}
	_, err := p.GetBinary(request)
	if err != nil {
		return err
	}
	m := gomail.NewMessage()
	p.setEmailHeader(m, *requestDTO)
	if err := p.setEmailBody(m, *requestDTO); err != nil {
		return err
	}
	if err := p.setEmailAttach(m, *requestDTO); err != nil {
		return err
	}
	smtpHost := requestDTO.SMTP.SmtpHost
	smtpPort := requestDTO.SMTP.SmtpPort
	smtpUser := requestDTO.SMTP.SmtpUser
	smtpPassword := requestDTO.SMTP.SmtpPassword
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("erro ao enviar o e-mail via Zoho: %w", err)
	}

	return nil
}

// reset generator resets the PDF generator to its initial state.
func (p *BillerMaroto) resetGenerator() {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Vertical).
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15).
		WithBottomMargin(15).
		Build()
	p.generator = maroto.New(cfg)
}

// setemailHeader sets the email header for sending the PDF.
func (p *BillerMaroto) setEmailHeader(m *gomail.Message, request dto.BillerRequest) {
	m.SetHeader("From", *request.Vendor.Email)
	m.SetHeader("To", *request.Customer.Email)
	m.SetHeader("Subject", p.getEmailSubject(request))
}

// setEmailBody sets the email body for sending .
func (p *BillerMaroto) setEmailBody(m *gomail.Message, request dto.BillerRequest) error {
	cidQRCode, err := p.prepareQRCodeAttachment(m, request)
	if err != nil {
		return err
	}
	body, err := p.getEmailBody(request, cidQRCode)
	if err != nil {
		return err
	}
	m.SetBody("text/html", body)
	return nil
}

// prepareQRCodeAttachment prepares the QR code attachment for the email.
func (p *BillerMaroto) prepareQRCodeAttachment(m *gomail.Message, request dto.BillerRequest) (string, error) {
	cidQRCode := "qrcode"
	if !strings.Contains(bodyTemplate, "{{.CidQRCode}}") {
		return "", nil
	}
	binQrCode, err := base64.StdEncoding.DecodeString(request.Receive.Pix.PixQRCode)
	if err != nil {
		return "", fmt.Errorf("failed to decode QR code base64: %w", err)
	}
	m.Attach("grafico.png", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(binQrCode)
		return err
	}), gomail.SetHeader(map[string][]string{
		"Content-Disposition": {`inline; filename="grafico.png"`},
		"Content-ID":          {fmt.Sprintf("<%s>", cidQRCode)}, // O ID precisa estar entre < >
	}))
	return cidQRCode, nil
}

// setEmailAttach sets the email attachment for sending the PDF.
func (p *BillerMaroto) setEmailAttach(m *gomail.Message, request dto.BillerRequest) error {
	m.Attach(request.BillFileName, gomail.SetCopyFunc(func(w io.Writer) error {
		bin, err := p.GetBinary(&request)
		if err != nil {
			return err
		}
		_, err = w.Write(bin)
		return err
	}))
	return nil
}

// getEmailSubject generates the email subject for sending the PDF.
func (p *BillerMaroto) getEmailSubject(request dto.BillerRequest) string {
	porMonth := map[string]string{
		"January":   "janeiro",
		"February":  "fevereiro",
		"March":     "março",
		"April":     "abril",
		"May":       "maio",
		"June":      "junho",
		"July":      "julho",
		"August":    "agosto",
		"September": "setembro",
		"October":   "outubro",
		"November":  "novembro",
		"December":  "dezembro",
	}
	month := request.InvoiceDate.Format("January")
	year := request.InvoiceDate.Format("2006")
	return fmt.Sprintf("Estúdio Amelia Cardoso - sua fatura de %s de %s ", porMonth[month], year)
}

// getEmailBody generates the email body for sending the PDF.
func (p *BillerMaroto) getEmailBody(request dto.BillerRequest, cidQRCode string) (string, error) {
	emailData := bodyData{
		VendorName:    request.Vendor.TradingName,
		Name:          strings.Split(request.Customer.Name, " ")[0],
		InvoiceDate:   request.InvoiceDate.Format("02/01/2006"),
		DueDate:       request.InvoiceDue.Format("02/01/2006"),
		Amount:        p.getEmailAmount(request),
		Description:   template.HTML(p.getEmailDescription(request)),
		QrCodeBase64:  template.HTML("data:image/png;base64," + request.Receive.Pix.PixQRCode),
		CidQRCode:     cidQRCode,
		CopyPasteCode: request.Receive.Pix.PixCopyPaste,
	}
	var body bytes.Buffer
	tmp := template.Must(template.New("email").Parse(bodyTemplate))
	if err := tmp.Execute(&body, emailData); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}
	return body.String(), nil
}

// getEmailDescription generates the email description for sending the PDF.
func (p *BillerMaroto) getEmailDescription(request dto.BillerRequest) string {
	var desc strings.Builder
	for _, item := range request.Items {
		desc.WriteString(fmt.Sprintf("<li>%d x %s</li>", item.Quantity, item.Description))
	}
	if desc.Len() > 0 {
		return "<ul>" + desc.String() + "</ul>"
	}
	return ""
}

// getEmailAmount gets the total amount for the email body.
func (p *BillerMaroto) getEmailAmount(request dto.BillerRequest) string {
	amount := 0.00
	for _, item := range request.Items {
		amount += float64(item.Quantity) * item.Price
	}
	amountStr := fmt.Sprintf("R$ %.2f", amount)
	amountStr = strings.Replace(amountStr, ".", ",", 1)
	return amountStr
}

// getForm gets biller form
func (p *BillerMaroto) getForm(request dto.BillerRequest) (core.Document, error) {
	p.resetGenerator()
	p.addFooter(request.Vendor.TradingName)
	p.addHeader(request)
	p.addSeparator(6, 3)
	p.addBill(request)
	p.addSeparator(4, 2)
	total := p.addItems(request)
	p.addSeparator(2, 2)
	p.addTotal(total)
	p.addSpaceRow(4)
	p.addReceive(request.Receive, total)
	p.addSpaceRow(4)
	p.addInstructions(request)
	p.addSpaceRow(8)
	p.addSignature(request)
	return p.generator.Generate()
}

// header
func (p *BillerMaroto) addHeader(request dto.BillerRequest) {
	p.generator.AddRow(20,
		image.NewFromFileCol(12, request.Vendor.Logo,
			props.Rect{
				Center:  true,
				Percent: 100,
			}))
	p.generator.AddRow(5,
		text.NewCol(12, request.Vendor.TradingName,
			props.Text{
				Top:   0,
				Style: fontstyle.Bold,
				Align: align.Center,
				Size:  12,
			}))
	p.generator.AddRow(5,
		text.NewCol(12, fmt.Sprintf("cnpj: %s", request.Vendor.Document),
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  10,
			}))
	if request.Vendor.Email != nil {
		p.generator.AddRow(5,
			text.NewCol(12, fmt.Sprintf("email: %s", *request.Vendor.Email),
				props.Text{
					Top:   0,
					Align: align.Center,
					Size:  10,
				}))
	}
	if request.Vendor.Whatsapp != nil {
		p.generator.AddRow(5,
			text.NewCol(12, fmt.Sprintf("whatsapp: %s", *request.Vendor.Whatsapp),
				props.Text{
					Top:   0,
					Align: align.Center,
					Size:  10,
				}))
	}
}

// instructions
func (p *BillerMaroto) addInstructions(request dto.BillerRequest) {
	txt := fmt.Sprintf("Em caso de dúvidas, entre em contato através do email: %s ou whatsapp: %s", *request.Vendor.Email, *request.Vendor.Whatsapp)
	p.generator.AddRow(5,
		text.NewCol(12, txt,
			props.Text{
				Left:  10,
				Top:   0,
				Align: align.Left,
				Size:  10,
			}))
}

// addSignature adds the signature section to the PDF.
func (p *BillerMaroto) addSignature(request dto.BillerRequest) {
	txt := fmt.Sprintf("%s", request.Vendor.TradingName)
	p.generator.AddRow(5,
		text.NewCol(12, txt,
			props.Text{
				Left:  10,
				Top:   0,
				Style: fontstyle.BoldItalic,
				Align: align.Left,
				Size:  10,
			}))
}

// addBill adds the bill information to the PDF.
func (p *BillerMaroto) addBill(request dto.BillerRequest) {
	p.generator.AddRow(5,
		text.NewCol(8, request.Customer.Name,
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		text.NewCol(5, p.getInvoiceID(request.InvoiceID),
			props.Text{
				Align: align.Left,
				Left:  6,
				Size:  10,
			}),
	)
	p.generator.AddRow(5,
		text.NewCol(8, p.getDocument(request.Customer.Document),
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  10,
			}),
		text.NewCol(5, p.getDueDate(request.InvoiceDue),
			props.Text{
				Align: align.Left,
				Left:  6,
				Style: fontstyle.Bold,
				Size:  10,
			}),
	)
	p.generator.AddRow(5,
		text.NewCol(8, p.getEmail(request.Customer.Email),
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  10,
			}),
		text.NewCol(5, p.getInvoiceDate(request.InvoiceDate),
			props.Text{
				Align: align.Left,
				Left:  6,
				Size:  10,
			}),
	)
}

// addSpaceRow adds empty rows to the PDF for spacing.
func (p *BillerMaroto) addSpaceRow(height float64) {
	p.generator.AddRow(height)
}

// addSeparator adds a separator row to the PDF.
func (p *BillerMaroto) addSeparator(heightTop float64, heightBottom float64) {
	p.generator.AddRow(heightTop)
	p.generator.AddRows(mline.NewRow(0, props.Line{
		Thickness: 0.5,
		Style:     linestyle.Solid,                               // Estilo (Solid, Dashed ou Dotted)
		Color:     &props.Color{Red: 200, Green: 200, Blue: 200}, // Cor vermelha
	}))
	p.generator.AddRow(heightBottom)
}

// addItems adds the items to the PDF.
func (p *BillerMaroto) addItems(request dto.BillerRequest) float64 {
	p.addItemHeader()
	var total float64
	for _, item := range request.Items {
		total += p.addItemRow(item)
	}
	return total
}

// getInvoiceID returns the invoice ID as a formatted string.
func (p *BillerMaroto) getInvoiceID(invoiceID int64) string {
	printer := message.NewPrinter(language.Portuguese)
	return printer.Sprintf("Invoice #      : %d", invoiceID)
}

// getDocument
func (p *BillerMaroto) getDocument(document *string) string {
	if document == nil {
		return ""
	}
	document_num := *document
	document_type := "CPF"
	reg := regexp.MustCompile(`\D`)
	if len(reg.ReplaceAllString(document_num, "")) >= 14 {
		document_type = "CNPJ"
	}
	return fmt.Sprintf("%s: %s", document_type, document_num)
}

// getInvoiceDate returns the invoice date as a formatted string.
func (p *BillerMaroto) getInvoiceDate(invoiceDate time.Time) string {
	return fmt.Sprintf("Emissão       : %s", invoiceDate.Format("02/01/2006"))
}

// getEmail returns the email as a formatted string.
func (p *BillerMaroto) getEmail(email *string) string {
	if email == nil {
		return ""
	}
	return fmt.Sprintf("Email: %s", *email)
}

// getDueDate returns the due date as a formatted string.
func (p *BillerMaroto) getDueDate(dueDate time.Time) string {
	return fmt.Sprintf("Vencimento: %s", dueDate.Format("02/01/2006"))
}

// addItemHeader adds the header for the items section in the PDF.
func (p *BillerMaroto) addItemHeader() {
	p.generator.AddRow(5,
		text.NewCol(6, "Descrição",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		text.NewCol(1, "Quantidade",
			props.Text{
				Top:   0,
				Align: align.Center,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		text.NewCol(2, "Valor",
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		text.NewCol(2, "Total",
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  10,
			}),
	)
}

// addItemRow adds a row for an item in the PDF.
func (p *BillerMaroto) addItemRow(item dto.BillerItem) float64 {
	price := fmt.Sprintf("R$ %.2f", item.Price)
	price = strings.Replace(price, ".", ",", 1)
	total := float64(item.Quantity) * item.Price
	totalStr := fmt.Sprintf("R$ %.2f", float64(item.Quantity)*item.Price)
	totalStr = strings.Replace(totalStr, ".", ",", 1)
	p.generator.AddRow(5,
		text.NewCol(6, item.Description,
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  10,
			}),
		text.NewCol(1, fmt.Sprintf("%d", item.Quantity),
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  10,
			}),
		text.NewCol(2, price,
			props.Text{
				Top:   0,
				Align: align.Right,
				Size:  10,
			}),
		text.NewCol(2, totalStr,
			props.Text{
				Top:   0,
				Align: align.Right,
				Size:  10,
			}),
	)
	return total
}

// addTotal
func (p *BillerMaroto) addTotal(total float64) {
	totalStr := fmt.Sprintf("R$ %.2f", total)
	totalStr = strings.Replace(totalStr, ".", ",", 1)
	p.generator.AddRow(5,
		text.NewCol(6, "",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  10,
			}),
		text.NewCol(1, "",
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  10,
			}),
		text.NewCol(2, "TOTAL",
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  10,
			}),
		text.NewCol(2, totalStr,
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  10,
			}),
	)
}

// addReceive adds the receive section to the PDF.
func (p *BillerMaroto) addReceive(receive dto.BillerReceive, value float64) error {
	if receive.BankAccount == nil && receive.Pix == nil {
		return nil
	}
	p.generator.AddRow(5,
		text.NewCol(12, "Notas:",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Style: fontstyle.Bold,
				Size:  10,
			}),
	)
	p.generator.AddRow(3)
	if err := p.addPix(receive.Pix, value); err != nil {
		return err
	}
	p.generator.AddRow(10)
	p.addBankAccount(receive.BankAccount, value)

	return nil
}

// addPix adds the Pix section to the PDF.
func (p *BillerMaroto) addPix(pix *dto.BillerPix, value float64) error {
	if pix == nil {
		return nil
	}
	p.generator.AddRow(5,
		text.NewCol(12, "Para pagamento via *Pix, utilize o QrCode ou as opções abaixo:",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  10,
			}),
	)

	cols := []core.Col{col.New(1)}

	qr, err := p.addQRCode(pix.PixQRCode)
	if err != nil {
		return err
	}
	if qr != nil {
		cols = append(cols, *qr)
	}

	cp := p.addCopyPaste(pix.PixCopyPaste, pix.PixKey, value)
	if cp != nil {
		cols = append(cols, *cp)
	}

	p.generator.AddRow(28, cols...)

	strVal := fmt.Sprintf("R$ %.2f", value)
	strVal = strings.Replace(strVal, ".", ",", -1)
	txt := fmt.Sprintf("* Por favor, antes de confirmar o pagamento, verifique que o valor é %s e que o recebedor é %s", strVal, pix.ReceiverName)

	p.generator.AddRow(3,
		text.NewCol(11, txt,
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)

	return nil
}

// addQRCode adds the QR code to the PDF.
func (p *BillerMaroto) addQRCode(qrCode string) (*core.Col, error) {
	img, err := base64.StdEncoding.DecodeString(qrCode)
	if err != nil {
		return nil, err
	}
	imgComp := image.NewFromBytes(img, "png")
	colr := col.New(2).Add(imgComp)

	return &colr, nil
}

// addCopyPaste adds the Pix copy-paste code to the PDF.
func (p *BillerMaroto) addCopyPaste(copyPaste string, pixKey string, value float64) *core.Col {
	strVal := fmt.Sprintf("valor R$ %.2f", value)
	strVal = strings.Replace(strVal, ".", ",", -1)

	pkey := fmt.Sprintf("%s (%s)", pixKey, strVal)
	colr := col.New(8).Add(
		text.New("Código Pix (Copie e Cole o código abaixo):", props.Text{Top: 1, Style: fontstyle.Bold, Align: align.Left, Left: 5, Size: 8}),
		text.New(copyPaste, props.Text{Top: 5, Align: align.Left, Left: 5, Size: 8, Color: &props.Color{Red: 0, Green: 0, Blue: 139}}),
		text.New("Chave Pix (caso queira pagar diretamente e enviar o comprovante):",
			props.Text{Top: 19, Style: fontstyle.Bold, Align: align.Left, Left: 5, Size: 8}),
		text.New(pkey, props.Text{Top: 23, Align: align.Left, Left: 5, Size: 8}),
	) // Fecha a colunaEsquerda
	return &colr
}

// addBankAccount adds the bank account information to the PDF.
func (p *BillerMaroto) addBankAccount(bankAccount *dto.BillerBankAccount, value float64) {
	if bankAccount == nil {
		return
	}
	p.generator.AddRow(5,
		text.NewCol(12, "Para pagamento via transferência bancária, utilize os dados abaixo:",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  10,
			}),
	)
	p.generator.AddRow(2)
	p.generator.AddRow(4,
		text.NewCol(12, fmt.Sprintf("Banco: %s", bankAccount.BankName),
			props.Text{
				Top:   0,
				Left:  17,
				Style: fontstyle.Bold,
				Align: align.Left,
				Size:  8,
			}),
	)
	p.generator.AddRow(4,
		text.NewCol(12, fmt.Sprintf("Agência: %s", bankAccount.BankAgency),
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)
	p.generator.AddRow(4,
		text.NewCol(12, fmt.Sprintf("Conta: %s", bankAccount.BankAccount),
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)
	p.generator.AddRow(4,
		text.NewCol(12, fmt.Sprintf("%s", bankAccount.ReceiverName),
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)
	p.generator.AddRow(4,
		text.NewCol(12, fmt.Sprintf("%s", bankAccount.ReceiverDocument),
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)
	strVal := fmt.Sprintf("Valor: R$ %.2f", value)
	strVal = strings.Replace(strVal, ".", ",", -1)
	p.generator.AddRow(4,
		text.NewCol(12, strVal,
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)
}

// addFooter adds the footer to the PDF.
func (p *BillerMaroto) addFooter(name string) {

	row0 := mline.NewRow(0, props.Line{
		Thickness: 0.5,
		Style:     linestyle.Solid,
		Color:     &props.Color{Red: 200, Green: 200, Blue: 200},
	})

	row1 := text.NewRow(
		5, "Este documento é uma representação do boleto e não possui validade fiscal.",
		props.Text{
			Top:   0,
			Align: align.Center,
			Size:  10,
		},
	)

	row2 := text.NewRow(
		5, "Gerado por: "+name,
		props.Text{
			Top:   0,
			Align: align.Center,
			Size:  10,
		},
	)

	p.generator.RegisterFooter(row0, row1, row2)
}
