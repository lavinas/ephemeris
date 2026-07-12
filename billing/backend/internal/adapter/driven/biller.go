package driven

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"billing/internal/dto"
	"billing/internal/port"

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
	p.addFooter(request.Vendor.Name)
	p.addHeader(request)
	p.addSeparator(6, 3)
	p.addBill(request)
	p.addSeparator(6, 2)
	total := p.addItems(request)
	p.addSeparator(2, 2)
	p.addTotal(total)
	p.addSpaceRow(5)
	p.addReceive(request.Receive, total)
	p.addSpaceRow(5)
	p.addInstructions(request)
	p.addSpaceRow(8)
	p.addSignature(request)
	
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
func (p *Biller) addInstructions(request dto.BillerRequest) {
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
func (p *Biller) addSignature(request dto.BillerRequest) {
	txt := fmt.Sprintf("%s", request.Vendor.Name)
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
func (p *Biller) addBill(request dto.BillerRequest) {
	p.generator.AddRow(4,
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
	p.generator.AddRow(4,
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
	p.generator.AddRow(4,
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
				Size:  10,
			}),
		text.NewCol(2, "Quantidade",
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
				Size:  12,
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
				Size:  10,
			}),
		text.NewCol(2, fmt.Sprintf("%d", item.Quantity),
			props.Text{
				Top:   0,
				Align: align.Center,
				Size:  10,
			}),
		text.NewCol(2, fmt.Sprintf("R$ %.2f", item.Price),
			props.Text{
				Top:   0,
				Align: align.Right,
				Size:  10,
			}),
		text.NewCol(2, fmt.Sprintf("R$ %.2f", total),
			props.Text{
				Top:   0,
				Align: align.Right,
				Size:  10,
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
				Size:  10,
			}),
		text.NewCol(2, "",
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
		text.NewCol(2, fmt.Sprintf("R$ %.2f", total),
			props.Text{
				Top:   0,
				Align: align.Right,
				Style: fontstyle.Bold,
				Size:  10,
			}),
	)
}

// addReceive adds the receive section to the PDF.
func (p *Biller) addReceive(receive dto.BillerReceive, value float64) error {
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
	p.generator.AddRow(13)
	p.addBankAccount(receive.BankAccount, value)

	return nil
}

// addPix adds the Pix section to the PDF.
func (p *Biller) addPix(pix *dto.BillerPix, value float64) error {
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


	txt := fmt.Sprintf("* Por favor, antes de confirmar o pagamento, verifique que o valor do pagamento é R$ %.2f e que o recebedor é %s", value, pix.ReceiverName)

	p.generator.AddRow(3,
		text.NewCol(11, txt,
			props.Text{
				Top:   10,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)

	return nil
}

// addQRCode adds the QR code to the PDF.
func (p *Biller) addQRCode(qrCode string) (*core.Col, error) {
	strImg := strings.SplitN(qrCode, ",", 2)[1]

	img, err := base64.StdEncoding.DecodeString(strImg)
	if err != nil {
		return nil, err
	}
	imgComp := image.NewFromBytes(img, "png")
	colr := col.New(2).Add(imgComp)
	return &colr, nil
}


// addCopyPaste adds the Pix copy-paste code to the PDF.
func (p *Biller) addCopyPaste(copyPaste string, pixKey string, value float64) *core.Col {
	pkey := fmt.Sprintf("%s (valor: R$ %.2f)", pixKey, value)
	colr := col.New(8).Add(
		text.New("Código Pix (Copie e Cole o código abaixo):", props.Text{Top: 4, Style: fontstyle.Bold, Align: align.Left, Left: 5, Size: 8}),
		text.New(copyPaste, props.Text{Top: 8, Align: align.Left, Left: 5, Size: 8}),
		text.New("Chave Pix (caso queira pagar diretamente e enviar o comprovante):", 
			props.Text{Top: 20, Style: fontstyle.Bold, Align: align.Left, Left: 5, Size: 8}),
		text.New(pkey, props.Text{Top: 24, Align: align.Left, Left: 5, Size: 8}),
	) // Fecha a colunaEsquerda
	return &colr
}

// addBankAccount adds the bank account information to the PDF.
func (p *Biller) addBankAccount(bankAccount *dto.BillerBankAccount, value float64) {
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
	p.generator.AddRow(4,
		text.NewCol(12, fmt.Sprintf("Valor: R$ %.2f", value),
			props.Text{
				Top:   0,
				Left:  17,
				Align: align.Left,
				Size:  8,
			}),
	)
}


// addFooter adds the footer to the PDF.
func (p *Biller) addFooter(name string) {

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
		5, "Gerado por: " + name,
		props.Text{
			Top:   0,
			Align: align.Center,
			Size:  10,
		},
	)

	p.generator.RegisterFooter(row0, row1, row2)
}
