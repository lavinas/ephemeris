package driven

import (
	"fmt"
	"regexp"
	"time"
	"encoding/base64"

	"billing/internal/dto"
	"billing/internal/port"

	"github.com/johnfercher/maroto/v2"
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
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
)

// PDFGenerator is an implementation of the PDFGenerator interface that generates PDF files.
type Biller struct {
	logger    port.Logger
	path      string
	generator core.Maroto
}

// NewBiller creates a new instance of Biller.
func NewBiller(logger port.Logger, path string) *Biller {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Vertical).
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15).
		WithBottomMargin(15).
		Build()
	return &Biller{
		logger:    logger,
		path:      path,
		generator: maroto.New(cfg),
	}
}

// GeneratePDF generates a PDF file based on the provided data and returns the file path.
func (p *Biller) GeneratePDF(request dto.BillerRequest) error {
	p.logger.IPrintf(2, "Generating PDF...")
	p.addHeader(request)
	p.addSeparator(3, 3)
	p.addBill(request)
	p.addSeparator(3, 2)
	total := p.addItems(request)
	p.addSeparator(2, 1)
	p.addTotal(total)
	p.addSpaceRow(5)
	p.addReceive(request.Receive)
	document, err := p.generator.Generate()
	if err != nil {
		return err
	}
	err = document.Save(p.path)
	if err != nil {
		return err
	}
	p.logger.IPrintf(2, "PDF generated successfully at path: %s", p.path)
	return nil
}

// header
func (p *Biller) addHeader(request dto.BillerRequest) {
	p.generator.AddRow(20,
		image.NewFromFileCol(12, request.Vendor.Logo,
			props.Rect{
				Center:  true,
				Percent: 100,
			}))
	p.generator.AddRow(5,
		text.NewCol(12, request.Vendor.Name,
			props.Text{
				Top:   0,
				Style: fontstyle.Bold,
				Align: align.Center,
				Size:  10,
			}))
	p.generator.AddRow(5,
		text.NewCol(12, fmt.Sprintf("cnpj: %s", request.Vendor.Document),
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  8,
			}))
	if request.Vendor.Email != nil {
		p.generator.AddRow(5,
			text.NewCol(12, fmt.Sprintf("email: %s", *request.Vendor.Email),
				props.Text{
					Top:   0,
					Align: align.Center,
					Size:  8,
				}))
	}
	if request.Vendor.Whatsapp != nil {
		p.generator.AddRow(5,
			text.NewCol(12, fmt.Sprintf("whatsapp: %s", *request.Vendor.Whatsapp),
				props.Text{
					Top:   0,
					Align: align.Center,
					Size:  8,
				}))
	}
}

// addBill adds the bill information to the PDF.
func (p *Biller) addBill(request dto.BillerRequest) {
	p.generator.AddRow(4,
		text.NewCol(8, request.Customer.Name,
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Style: fontstyle.Bold,
				Size:  8,
			}),
		text.NewCol(5, p.getInvoiceID(request.InvoiceID),
			props.Text{
				Align: align.Left,
				Left:  15,
				Size:  8,
			}),
	)
	p.generator.AddRow(4,
		text.NewCol(8, p.getDocument(request.Customer.Document),
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  8,
			}),
		text.NewCol(5, p.getDueDate(request.InvoiceDue),
			props.Text{
				Align: align.Left,
				Left:  15,
				Style: fontstyle.Bold,
				Size:  8,
			}),
	)
	p.generator.AddRow(4,
		text.NewCol(8, p.getEmail(request.Customer.Email),
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  8,
			}),
		text.NewCol(5, p.getInvoiceDate(request.InvoiceDate),
			props.Text{
				Align: align.Left,
				Left:  15,
				Size:  8,
			}),
	)
}

// addSpaceRow adds empty rows to the PDF for spacing.
func (p *Biller) addSpaceRow(height float64) {
	p.generator.AddRow(height)
}

// addSeparator adds a separator row to the PDF.
func (p *Biller) addSeparator(heightTop float64, heightBottom float64) {
	p.generator.AddRow(heightTop)
	p.generator.AddRows(mline.NewRow(0, props.Line{
		Thickness: 0.5,
		Style:     linestyle.Solid,                               // Estilo (Solid, Dashed ou Dotted)
		Color:     &props.Color{Red: 200, Green: 200, Blue: 200}, // Cor vermelha
	}))
	p.generator.AddRow(heightBottom)
}

// addItems adds the items to the PDF.
func (p *Biller) addItems(request dto.BillerRequest) float64 {
	p.addItemHeader()
	var total float64
	for _, item := range request.Items {
		total += p.addItemRow(item)
	}
	return total
}

// getInvoiceID returns the invoice ID as a formatted string.
func (p *Biller) getInvoiceID(invoiceID string) string {
	return fmt.Sprintf("Invoice #      : %s", invoiceID)
}

// getDocument
func (p *Biller) getDocument(document *string) string {
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
func (p *Biller) getInvoiceDate(invoiceDate time.Time) string {
	return fmt.Sprintf("Emissão       : %s", invoiceDate.Format("02/01/2006"))
}

// getEmail returns the email as a formatted string.
func (p *Biller) getEmail(email *string) string {
	if email == nil {
		return ""
	}
	return fmt.Sprintf("Email: %s", *email)
}

// getDueDate returns the due date as a formatted string.
func (p *Biller) getDueDate(dueDate time.Time) string {
	return fmt.Sprintf("Vencimento: %s", dueDate.Format("02/01/2006"))
}

// addItemHeader adds the header for the items section in the PDF.
func (p *Biller) addItemHeader() {
	p.generator.AddRow(5,
		text.NewCol(5, "Descrição",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Style: fontstyle.Bold,
				Size:  8,
			}),
		text.NewCol(2, "Quantidade",
			props.Text{
				Top:   0,
				Align: align.Center,
				Style: fontstyle.Bold,
				Size:  8,
			}),
		text.NewCol(2, "Valor",
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  8,
			}),
		text.NewCol(2, "Total",
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  8,
			}),
	)
}

// addItemRow adds a row for an item in the PDF.
func (p *Biller) addItemRow(item dto.BillerItem) float64 {
	total := float64(item.Quantity) * item.Price
	p.generator.AddRow(5,
		text.NewCol(5, item.Description,
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  8,
			}),
		text.NewCol(2, fmt.Sprintf("%d", item.Quantity),
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  8,
			}),
		text.NewCol(2, fmt.Sprintf("R$ %.2f", item.Price),
			props.Text{
				Top:   0,
				Align: align.Right,
				Size:  8,
			}),
		text.NewCol(2, fmt.Sprintf("R$ %.2f", total),
			props.Text{
				Top:   0,
				Align: align.Right,
				Size:  8,
			}),
	)
	return total
}

// addTotal
func (p *Biller) addTotal(total float64) {
	p.generator.AddRow(5,
		text.NewCol(5, "",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Size:  8,
			}),
		text.NewCol(2, "",
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  8,
			}),
		text.NewCol(2, "TOTAL",
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  8,
			}),
		text.NewCol(2, fmt.Sprintf("R$ %.2f", total),
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  8,
			}),
	)
}

// addReceive adds the receive section to the PDF.
func (p *Biller) addReceive(receive dto.BillerReceive) {
	if receive.BankAccount == nil && receive.Pix == nil {
		return
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
	p.generator.AddRow(5)
	p.addPix(receive.Pix)
	
}

// addPix adds the Pix section to the PDF.
func (p *Biller) addPix(pix *dto.BillerPix) {
	if pix == nil {
		return
	}
	p.generator.AddRow(5,
		text.NewCol(12, "Para pagamento via Pix:",
			props.Text{
				Top:   0,
				Left:  10,
				Align: align.Left,
				Style: fontstyle.Bold,
				Size:  10,
			}),
	)

	p.generator.AddRow(5)
	p.addQRCode(pix.PixQRCode)

}

// addQRCode adds the QR code to the PDF.
func (p *Biller) addQRCode(qrCode *string) error {
	if qrCode == nil {
		return nil
	}
	img, err := base64.StdEncoding.DecodeString(*qrCode)
	if err != nil {
		return err
	}
	imgComp := image.NewFromBytes(img, "png")

	p.generator.AddRows(
		row.New(40).Add(
			col.New(12).Add(imgComp),
		),
	)
	return nil
}


