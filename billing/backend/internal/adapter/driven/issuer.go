package driven

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"billing/internal/dto"
	"billing/internal/port"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"gopkg.in/gomail.v2"
)

const (
	generateTimeout = 15 * time.Second
	paperWidth      = 8.27  // A4 width in inches
	paperHeight     = 11.69 // A4 height in inches
	marginTop       = 0.59  // Top margin (~1.5 cm)
	marginBottom    = 0.59  // Bottom margin (~1.5 cm)
	marginLeft      = 0.59  // Left margin (~1.5 cm)
	marginRight     = 0.59  // Right margin (~1.5 cm)
)

// Issuer is responsible for handling the receipt of emissions.
type Issuer struct {
}

// NewIssuer creates a new instance of Issuer.
func NewIssuer() *Issuer {
	return &Issuer{}
}

// GetBase64 generates a receipt for the given data and returns it as a base64-encoded PDF.
func (r *Issuer) GetBase64(data port.InDTO, html_pdf string) ([]byte, error) {
	dtoData, ok := data.(*dto.IssuerData)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected *dto.IssuerData")
	}
	if dtoData.VendorLogoBase64 != "" && !strings.HasPrefix(string(dtoData.VendorLogoBase64), "data:image") {
		dtoData.VendorLogoBase64 = template.URL("data:image/png;base64," + string(dtoData.VendorLogoBase64))
	}
	htmlContent, err := r.format(*dtoData, html_pdf, "", "")
	if err != nil {
		return nil, fmt.Errorf("GetBase64: %w", err)
	}
	pdf, err := r.getPDF(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("GetBase64: %w", err)
	}
	return pdf, nil
}

// SendMail sends the generated receipt via email using the provided SMTP configuration.
func (r *Issuer) SendMail(data port.InDTO, subject string, html_pdf string, html_email string) error {
	dtoData, ok := data.(*dto.IssuerData)
	if !ok {
		return fmt.Errorf("invalid data type: expected *dto.IssuerData")
	}
	if dtoData.CustomerEmail == "" {
		return fmt.Errorf("customer email is required")
	}
	pdfBase64, err := r.GetBase64(dtoData, html_pdf)
	if err != nil {
		return fmt.Errorf("SendMail: %w", err)
	}
	logoImage, logoAlias, err := r.emailLogoAttachment(dtoData, html_email)
	if err != nil {
		return fmt.Errorf("SendMail: %w", err)
	}
	qrImage, qrAlias, err := r.emailQRCodeAttachment(dtoData, html_email)
	if err != nil {
		return fmt.Errorf("SendMail: %w", err)
	}

	htmlContent, err := r.format(*dtoData, html_email, logoAlias, qrAlias)
	if err != nil {
		return fmt.Errorf("SendMail: %w", err)
	}
	return r.send(*dtoData, subject, htmlContent, pdfBase64, logoImage, logoAlias, qrImage, qrAlias)
}

// GetReceiptName generates a filename for the receipt PDF based on the invoice number and customer name.
func (r *Issuer) GetName(data port.InDTO) string {
	dtoData, ok := data.(*dto.IssuerData)
	if !ok {
		return "receipt.pdf"
	}
	return fmt.Sprintf("%s-%s-%s-recibo.pdf", dtoData.InvoiceDueDate.Format("2006-01-02"),
		dtoData.InvoiceDate.Format("2006-01-02"), dtoData.CustomerNickname)
}

// prepareLogoAttachment prepares the logo attachment for the email.
func (r *Issuer) emailLogoAttachment(dtoData *dto.IssuerData, html_email string) ([]byte, string, error) {
	alias := "logo"
	if dtoData.VendorLogoBase64 == "" {
		return nil, "", nil // No logo provided
	}
	if !strings.Contains(html_email, ".VendorLogoBase64") {
		return nil, "", nil // Logo not referenced in the email template
	}
	img := string(dtoData.VendorLogoBase64)
	if strings.HasPrefix(img, "data:image") {
		parts := strings.SplitN(img, ",", 2)
		if len(parts) == 2 {
			img = parts[1]
		}
	}
	logoBytes, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		return nil, "", fmt.Errorf("decoding logo base64: %w", err)
	}
	return logoBytes, alias, nil
}

// prepareQRCodeAttachment prepares the QR code attachment for the email.
func (r *Issuer) emailQRCodeAttachment(dtoData *dto.IssuerData, html_email string) ([]byte, string, error) {
	alias := "qrcode"
	if dtoData.VendorPixQRBase64 == "" {
		return nil, "", nil // No QR code provided
	}
	if !strings.Contains(html_email, ".VendorPixQRBase64") {
		return nil, "", nil // QR code not referenced in the email template
	}
	img := string(dtoData.VendorPixQRBase64)
	if strings.HasPrefix(img, "data:image") {
		parts := strings.SplitN(img, ",", 2)
		if len(parts) == 2 {
			img = parts[1]
		}
	}
	qrCodeBytes, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		return nil, "", fmt.Errorf("decoding QR code base64: %w", err)
	}
	return qrCodeBytes, alias, nil
}

// send sends the generated receipt via email using the provided SMTP configuration.
func (r *Issuer) send(dtoData dto.IssuerData, subject string, htmlContent string, pdfBase64 []byte,
	logoImage []byte, logoAlias string, qrImage []byte, qrAlias string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", dtoData.VendorEmail)
	m.SetHeader("To", dtoData.CustomerEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlContent)
	m.Attach(r.GetName(&dtoData), gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(pdfBase64)
		return err
	}))
	if logoImage != nil {
		m.Attach("grafico.png", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(logoImage)
			return err
		}), gomail.SetHeader(map[string][]string{
			"Content-Disposition": {`inline; filename="grafico.png"`},
			"Content-ID":          {fmt.Sprintf("<%s>", logoAlias)}, // O ID precisa estar entre < >
		}))
	}
	if qrImage != nil {
		m.Attach("qrcode.png", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(qrImage)
			return err
		}), gomail.SetHeader(map[string][]string{
			"Content-Disposition": {`inline; filename="qrcode.png"`},
			"Content-ID":          {fmt.Sprintf("<%s>", qrAlias)}, // O ID precisa estar entre < >
		}))
	}
	d := gomail.NewDialer(dtoData.VendorSMTPHost, dtoData.VendorSMTPPort,
		dtoData.VendorSMTPUsername, dtoData.VendorSMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("SendMail: %w", err)
	}
	return nil
}

// getHTML generates the HTML content for the receipt based on the provided data.
func (r *Issuer) format(data dto.IssuerData, html string, logoAlias, qrAlias string) (string, error) {
	funcMap := template.FuncMap{
		"currency": func(amount float64) string {
			val := fmt.Sprintf("%.2f", amount)
			val = strings.Replace(val, ".", ",", 1)
			return val
		},
		"br_currency": func(amount float64) string {
			val := fmt.Sprintf("%.2f", amount)
			val = strings.Replace(val, ".", ",", 1)
			return "R$ " + val
		},
		"br_date": func(date time.Time) string {
			return date.Format("02/01/2006")
		},
	}
	html_tmpl, err := template.New("receipt").Funcs(funcMap).Parse(html)
	if err != nil {
		return "", fmt.Errorf("GeneratingReceiptBase64: %w", err)
	}
	if logoAlias != "" {
		data.VendorLogoBase64 = template.URL(fmt.Sprintf("cid:%s", logoAlias))
	}
	if qrAlias != "" {
		data.VendorPixQRBase64 = template.URL(fmt.Sprintf("cid:%s", qrAlias))
	}
	var buf bytes.Buffer
	err = html_tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("Executing template: %w", err)
	}
	return buf.String(), nil
}

// getPDF generates a PDF file for the given receipt data and saves it to the specified output path.
func (r *Issuer) getPDF(htmlContent string) ([]byte, error) {
	// start chromedp with custom options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Headless,
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	// create a new context with a timeout
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()
	ctx, cancelTimeout := context.WithTimeout(ctx, generateTimeout)
	defer cancelTimeout()
	// generate the PDF
	var pdfBuffer []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuffer, _, err = page.PrintToPDF().
				WithPaperWidth(paperWidth).
				WithPaperHeight(paperHeight).
				WithMarginTop(marginTop).
				WithMarginBottom(marginBottom).
				WithMarginLeft(marginLeft).
				WithMarginRight(marginRight).
				WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("Generating PDF: %w", err)
	}
	// return the generated PDF buffer
	return pdfBuffer, nil
}
