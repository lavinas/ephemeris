package driven

import (
	"billing/internal/dto"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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

// Receipter is responsible for handling the receipt of emissions.
type Receipter struct {
	receiptTemplate string
}

// NewReceipter creates a new instance of Receipter.
func NewReceipter(receiptPath string) (*Receipter, error) {
	file, err := os.ReadFile(receiptPath)
	if err != nil {
		return nil, err
	}
	receiptTemplate := string(file)
	return &Receipter{
		receiptTemplate: receiptTemplate,
	}, nil
}

// GenerateReceipt generates a receipt for the given data.
func (r *Receipter) GetReceiptBase64(data dto.ReceiptData) ([]byte, error) {
	htmlContent, err := r.getHTML(data)
	if err != nil {
		return nil, fmt.Errorf("GeneratingReceiptBase64: %w", err)
	}
	pdf, err := r.getPDF(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("GeneratingReceiptBase64: %w", err)
	}
	return pdf, nil
}

// getHTML generates the HTML content for the receipt based on the provided data.
func (r *Receipter) getHTML(data dto.ReceiptData) (string, error) {
	funcMap := template.FuncMap{
		"currency": func(amount float64) string {
			return fmt.Sprintf("%.2f", amount)
		},
		"br_currency": func(amount float64) string {
			return fmt.Sprintf("R$ %.2f", amount)
		},
		"br_date": func(date string) string {
			parsedDate, err := time.Parse("2006-01-02", date)
			if err != nil {
				return date // Return the original string if parsing fails
			}
			return parsedDate.Format("02/01/2006")
		},
	}
	html_tmpl, err := template.New("receipt").Funcs(funcMap).Parse(r.receiptTemplate)
	if err != nil {
		return "", fmt.Errorf("GeneratingReceiptBase64: %w", err)
	}
	var buf bytes.Buffer
	err = html_tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("Executing template: %w", err)
	}
	return buf.String(), nil
}

// getPDF generates a PDF file for the given receipt data and saves it to the specified output path.
func (r *Receipter) getPDF(htmlContent string) ([]byte, error) {
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
