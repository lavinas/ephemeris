package driven

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"billing/internal/domain"
	"billing/internal/port"
	"github.com/ianlopshire/go-fixedwidth"

	"golang.org/x/text/encoding/charmap"
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

// IssuerFile is a concrete implementation of the port.
type IssuerFile struct {
	filePath string
	pattern  string
	logger   port.Logger
	file     *os.File
	writer   *transform.Writer
}

// NewIssuerFile creates a new instance of IssuerFile with the specified file path and logger.
func NewIssuerFile(filePath string, filePattern string, logger port.Logger) *IssuerFile {
	return &IssuerFile{
		filePath: filePath,
		pattern:  filePattern,
		logger:   logger,
	}
}

// SendEmission sends the emission data to a file and logs the operation.
func (i *IssuerFile) SendEmission(emission *domain.Emission) error {
	i.logger.IPrintf(2, "Sending emission to file: %s", i.filePath)
	if err := i.openSendFile(emission); err != nil {
		return err
	}
	defer i.file.Close()
	if err := i.writeHeader(emission); err != nil {
		return err
	}
	if err := i.writeItems(emission); err != nil {
		return err
	}
	if err := i.writeFooter(emission); err != nil {
		return err
	}
	i.logger.IPrintf(2, "Emission ID: %d, Quantity: %d, Amount: %.2f", emission.ID, emission.Quantity, emission.Amount)
	return nil
}

// ReceiveEmission is a placeholder for receiving emissions.
func (i *IssuerFile) ReceiveEmission(source string) (map[int64]*domain.EmissionItem, error) {
	i.logger.IPrintf(2, "Receiving emission from file: %s", source)
	if err := i.openReceiveFile(source); err != nil {
		return nil, err
	}
	defer i.file.Close()
	lines, err := i.readReceiveFile()
	if err != nil {
		return nil, err
	}
	i.logger.IPrintf(2, "Received %d lines from file: %s", len(lines), source)
	return lines, nil
}

// openSendFile is a helper function to open the file for writing.
func (i *IssuerFile) openSendFile(emission *domain.Emission) error {
	file_path := filepath.Join(i.filePath, i.replacePlaceholders(i.pattern, emission))
	i.logger.IPrintf(3, "Opened file: %s (path: %s, pattern: %s)", file_path, i.filePath, i.pattern)
	file, err := os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	writer := transform.NewWriter(file, charmap.ISO8859_1.NewEncoder())

	i.file = file
	i.writer = writer

	return nil
}

// openReceiveFile is a placeholder for opening the file for reading.
func (i *IssuerFile) openReceiveFile(source string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	i.file = file
	return nil
}

// readReceiveFile is a placeholder for reading the file for receiving emissions.
func (i *IssuerFile) readReceiveFile() (map[int64]*domain.EmissionItem, error) {
	lines := make(map[int64]*domain.EmissionItem)
	reader := csv.NewReader(i.file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		i.logger.IPrintf(3, "Error on reading CSV file: %v", err)
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
func (i *IssuerFile) getItemFromRecord(record []string) (*domain.EmissionItem, error) {
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
	nfeVerification := record[3]
	item := &domain.EmissionItem{
		RPSNumber:       rpsNum,
		NFENumber:       &nfeNum,
		NFEDatetime:     &nfeDateTime,
		NFEVerification: &nfeVerification,
	}
	return item, nil
}

// replacePlaceholders replaces placeholders in the file pattern with actual values from the emission.
func (i *IssuerFile) replacePlaceholders(pattern string, emission *domain.Emission) string {
	year := emission.EmissionDate.Format("2006") // Get the year from the emission date
	month := emission.EmissionDate.Format("01")  // Get the month from the emission date
	day := emission.EmissionDate.Format("02")    // Get the day from the emission date
	id := strconv.FormatInt(emission.ID, 10)     // Convert the emission ID to a string
	pattern = strings.ReplaceAll(pattern, "<yyyy>", year)
	pattern = strings.ReplaceAll(pattern, "<mm>", month)
	pattern = strings.ReplaceAll(pattern, "<dd>", day) // Get the day from the emission date
	pattern = strings.ReplaceAll(pattern, "<id>", id)
	return pattern
}

// writeHeader writes the header to the file.
func (i *IssuerFile) writeHeader(emission *domain.Emission) error {
	i.logger.IPrintf(3, "Writing header for emission ID: %d", emission.ID)
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
	line := string(h)
	line = strings.TrimRight(line, " ") + "\n"
	if _, err := i.writer.Write([]byte(line)); err != nil {
		return err
	}

	i.logger.IPrintf(3, "Header written for emission ID: %d", emission.ID)
	return nil
}

// writelines writes the emission lines to the file.
func (i *IssuerFile) writeItems(emission *domain.Emission) error {
	i.logger.IPrintf(3, "Writing items for emission ID: %d", emission.ID)
	emissionDate, _ := strconv.Atoi(emission.EmissionDate.Format("20060102"))
	for _, item := range emission.EmissionItems {
		it := i.getSendLine(emissionDate, item)
		line, err := fixedwidth.Marshal(it)
		lineStr := string(line)
		lineStr = strings.TrimRight(lineStr, " ") + "\n"
		if err != nil {
			return err
		}
		if _, err := i.writer.Write([]byte(lineStr)); err != nil {
			return err
		}
		i.logger.IPrintf(3, "Item written for emission ID: %d, RPS Number: %d", emission.ID, item.RPSNumber)
	}
	i.logger.IPrintf(3, "All items written for emission ID: %d", emission.ID)
	return nil
}

// getSendLine is a helper function to convert an EmissionItem to a line.
func (i *IssuerFile) getSendLine(emissionDate int, item *domain.EmissionItem) line {
	// Convert item fields to the appropriate types and formats
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
	}
}

// getSendDescription is a helper function to generate a description for the emission item.
func (i *IssuerFile) getSendDescription(item *domain.EmissionItem) string {
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

// writeFooter writes the footer to the file. This is a placeholder implementation
func (i *IssuerFile) writeFooter(emission *domain.Emission) error {
	i.logger.IPrintf(3, "Writing footer for emission ID: %d", emission.ID)
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
	line := string(f)
	line = strings.TrimRight(line, " ") + "\n"
	if _, err := i.writer.Write([]byte(line)); err != nil {
		return err
	}
	i.logger.IPrintf(3, "Footer written for emission ID: %d", emission.ID)
	return nil
}

// removeAccents is a helper function to remove accents from a string.
func (i *IssuerFile) removeAccents(texto string) string {
	// Transforma a string para separar letras dos acentos, remove os acentos e recompõe
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// Aplica a transformação
	resultado, _, _ := transform.String(t, texto)

	return resultado
}
