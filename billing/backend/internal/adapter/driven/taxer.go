package driven

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"billing/internal/domain"
	"github.com/ianlopshire/go-fixedwidth"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	version        = 1
	serie          = "A"
	rps_type       = "RPS"
	rps_situation  = "T"
	service_code   = 5762
	aliquot        = 0
	retention_type = 2
)

// header
type header struct {
	RegType   int `fixed:"1,1,right,0"`
	Version   int `fixed:"2,4,right,0"`
	CCM       int `fixed:"5,12,right,0"`
	StartDate int `fixed:"13,20,right,0"`
	EndDate   int `fixed:"21,28,right,0"`
}

// item represents an individual emission item in the file.
type line struct {
	RegType           int    `fixed:"1,1,right,0"`
	RPSType           string `fixed:"2,6,left, "`
	Serie             string `fixed:"7,11,left, "`
	RPSNumber         int64  `fixed:"12,23,right,0"`
	EmissionDate      int    `fixed:"24,31,right,0"`
	Situation         string `fixed:"32,32,left, "`
	Amount            int    `fixed:"33,47,right,0"`
	Discount          int    `fixed:"48,62,right,0"`
	ServiceCode       int    `fixed:"63,67,right,0"`
	Aliquot           int    `fixed:"68,71,right,0"`
	RetentionType     int    `fixed:"72,72,right,0"`
	DocumentType      int    `fixed:"73,73,right,0"`
	Document          int    `fixed:"74,87,right,0"`
	CityDocument      int    `fixed:"88,95,right,0"`
	StateDocument     int    `fixed:"96,107,right,0"`
	Name              string `fixed:"108,182,left, "`
	AddressType       string `fixed:"183,185,left, "`
	Address           string `fixed:"186,235,left, "`
	AddressNumber     string `fixed:"236,245,left, "`
	AddressComplement string `fixed:"246,275,left, "`
	Neighborhood      string `fixed:"276,305,left, "`
	City              string `fixed:"306,355,left, "`
	State             string `fixed:"356,357,left, "`
	PostalCode        string `fixed:"358,365,left, "`
	Email             string `fixed:"366,440,left, "`
	Description       string `fixed:"441,1441,left, "`
}

// footer represents the footer of the emission file.
type footer struct {
	RegType       int `fixed:"1,1,right,0"`
	TotalRecords  int `fixed:"2,8,right,0"`
	TotalAmount   int `fixed:"9,23,right,0"`
	TotalDiscount int `fixed:"24,38,right,0"`
}

// Taxer is a concrete implementation of the port.
type Taxer struct {
}

// NewTaxer creates a new instance of Taxer with the specified file path and logger.
func NewTaxer() *Taxer {
	return &Taxer{}
}

// GetContent generates the content of the emission file as a string.
func (i *Taxer) GetEmission(emission *domain.Emission, builder *strings.Builder) error {
	if err := i.getHeaderLine(emission, builder); err != nil {
		return err
	}
	if err := i.getItems(emission, builder); err != nil {
		return err
	}
	if err := i.getFooter(emission, builder); err != nil {
		return err
	}
	return nil
}

// ReceiveEmission is a placeholder for receiving emissions.
func (i *Taxer) ClearEmission(source string) (map[int64]*domain.EmissionItem, error) {
	decoded, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(decoded)
	lines, err := i.readReceiveFile(reader)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

// readReceiveFile is a placeholder for reading the file for receiving emissions.
func (i *Taxer) readReceiveFile(reader *bytes.Reader) (map[int64]*domain.EmissionItem, error) {
	lines := make(map[int64]*domain.EmissionItem)
	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = -1
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	header := records[0]
	if header[0] != "Tipo de Registro" {
		return nil, fmt.Errorf("invalid file")
	}
	var totalTrailer int
	for _, record := range records[1:] {
		switch record[0] {
		case "2":
			item, err := i.getItemFromRecord(record)
			if err != nil {
				return nil, err
			}
			lines[item.RPSNumber] = item
		case "Total":
			totalTrailer, err = strconv.Atoi(record[1])
			if err != nil {
				return nil, fmt.Errorf("invalid trailer total: %v", err)
			}
		}
	}
	if totalTrailer != len(lines) {
		return nil, fmt.Errorf("trailer total does not match number of lines: %d != %d",
			totalTrailer, len(lines))
	}
	return lines, nil
}

// getItemFromRecord is a helper function to convert a CSV record to an EmissionItem.
func (i *Taxer) getItemFromRecord(record []string) (*domain.EmissionItem, error) {
	rpsNum, err := strconv.ParseInt(record[6], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid RPS number: %v", err)
	}
	nfeNum, err := strconv.ParseInt(record[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid NFE number: %v", err)
	}
	nfeDateTime, err := time.Parse("02/01/2006 15:04:05", record[2])
	if err != nil {
		return nil, fmt.Errorf("invalid NFE datetime: %v", err)
	}
	amount, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid NFE amount: %v", err)
	}
	nfeVerification := record[3]
	item := &domain.EmissionItem{
		RPSNumber:       rpsNum,
		NFENumber:       &nfeNum,
		NFEDatetime:     &nfeDateTime,
		NFEVerification: &nfeVerification,
		NFEAmount:       &amount,
	}
	return item, nil
}

// getHeaderLine is a helper function to convert an Emission to a header line.
func (i *Taxer) getHeaderLine(emission *domain.Emission, builder *strings.Builder) error {
	ccm := regexp.MustCompile(`[^0-9]`).ReplaceAllString(emission.Vendor.TaxDocument, "")
	ccmd, _ := strconv.Atoi(ccm)
	ed := emission.EmissionDate.Format("20060102")
	edd, _ := strconv.Atoi(ed)
	header := header{
		RegType:   1,
		Version:   version,
		CCM:       ccmd,
		StartDate: edd,
		EndDate:   edd,
	}
	h, err := fixedwidth.Marshal(header)
	if err != nil {
		return err
	}
	builder.WriteString(strings.TrimRight(string(h), " ") + "\n")
	return nil
}

// getItems is a helper function to convert an Emission to a slice of lines.
func (i *Taxer) getItems(emission *domain.Emission, builder *strings.Builder) error {
	emissionDate, _ := strconv.Atoi(emission.EmissionDate.Format("20060102"))
	for _, item := range emission.EmissionItems {
		it, err := i.getSendLine(emissionDate, item)
		if err != nil {
			return err
		}
		line, err := fixedwidth.Marshal(it)
		if err != nil {
			return err
		}
		lineStr := string(line)
		lineStr = strings.TrimRight(lineStr, " ") + "\n"
		builder.WriteString(lineStr)
	}
	return nil
}

// getSendLine is a helper function to convert an EmissionItem to a line.
func (i *Taxer) getSendLine(emissionDate int, item *domain.EmissionItem) (line, error) {
	// Convert item fields to the appropriate types and formats
	if item.Invoice.Customer.Document == nil {
		return line{}, fmt.Errorf("missing customer document for RPS number: %d", item.RPSNumber)
	}
	document := regexp.MustCompile(`[^0-9]`).ReplaceAllString(*item.Invoice.Customer.Document, "")
	documentType := 1
	if len(document) > 11 {
		documentType = 2
	}
	docnumber, _ := strconv.Atoi(document)
	email := " "
	if item.Invoice.Customer.Email != nil {
		email = *item.Invoice.Customer.Email
	}
	name := i.removeAccents(item.Invoice.Customer.Name)
	return line{
		RegType:           2,
		RPSType:           rps_type,
		Serie:             serie,
		RPSNumber:         item.RPSNumber,
		EmissionDate:      emissionDate,
		Situation:         rps_situation,
		Amount:            int(item.Invoice.Amount * 100),
		Discount:          0,
		ServiceCode:       service_code,
		Aliquot:           aliquot,
		RetentionType:     retention_type,
		DocumentType:      documentType,
		Document:          docnumber,
		CityDocument:      0,
		StateDocument:     0,
		Name:              name,
		AddressType:       " ",
		Address:           " ",
		AddressNumber:     " ",
		AddressComplement: " ",
		Neighborhood:      " ",
		City:              " ",
		State:             " ",
		PostalCode:        " ",
		Email:             email,
		Description:       i.getSendDescription(item),
	}, nil
}

// getSendDescription is a helper function to generate a description for the emission item.
func (i *Taxer) getSendDescription(item *domain.EmissionItem) string {
	// Implement logic to generate a description based on the item details
	ret := ""
	for _, invItem := range item.Invoice.InvoiceItems {
		ret += strconv.Itoa(invItem.Quantity) + " " + invItem.Description + "|"
	}
	if len(ret) > 0 {
		ret = ret[:len(ret)-1] // Remove the last "|"
	}
	return ret
}

// getFooter is a helper function to convert an Emission to a footer line. This is a placeholder implementation
func (i *Taxer) getFooter(emission *domain.Emission, builder *strings.Builder) error {
	amount := int(emission.Amount * 100)
	footer := footer{
		RegType:       9,
		TotalRecords:  len(emission.EmissionItems),
		TotalAmount:   amount,
		TotalDiscount: 0,
	}
	f, err := fixedwidth.Marshal(footer)
	if err != nil {
		return err
	}
	builder.WriteString(strings.TrimRight(string(f), " ") + "\n")
	return nil
}

// removeAccents is a helper function to remove accents from a string.
func (i *Taxer) removeAccents(texto string) string {
	// Transforma a string para separar letras dos acentos, remove os acentos e recompõe
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// Aplica a transformação
	resultado, _, _ := transform.String(t, texto)

	return resultado
}
