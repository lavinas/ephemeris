package driven

import (
	"billing/internal/dto"
	"billing/internal/port"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
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
	p.addHeader(request)
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
			},
		),
	)
	p.generator.AddRow(10,
		text.NewCol(12, request.Vendor.Name,
			props.Text{
				Top:   0,
				Style: fontstyle.Bold,
				Align: align.Center,
				Size:  10,
			}))
	p.generator.AddRow(10,
		text.NewCol(12, "Invoice", props.Text{
			Top:   5,
			Style: fontstyle.Bold,
			Align: align.Center,
			Size:  12,
		}),
	)
}
