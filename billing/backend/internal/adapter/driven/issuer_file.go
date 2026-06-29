package driven

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"billing/internal/domain"
	"billing/internal/port"
	"github.com/ianlopshire/go-fixedwidth"
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
	Version   int `fixed:"2,3,right,0"`
	CCM       int `fixed:"5,8,right,0"`
	StartDate int `fixed:"13,8,right,0"`
	EndDate   int `fixed:"21,8,right,0"`
}

// item represents an individual emission item in the file.
type line struct {
	RegType           int    `fixed:"1,1,right,0"`
	RPSType           string `fixed:"2,5,left, "`
	Serie             string `fixed:"7,5,left, "`
	RPSNumber         int64  `fixed:"12,12,right,0"`
	EmissionDate      int    `fixed:"24,8,right,0"`
	Situation         string `fixed:"32,1,left, "`
	Amount            int    `fixed:"33,15,right,0"`
	Discount          int    `fixed:"48,15,right,0"`
	ServiceCode       int    `fixed:"63,5,right,0"`
	Aliquot           int    `fixed:"68,4,right,0"`
	RetentionType     int    `fixed:"72,1,right,0"`
	DocumentType      int    `fixed:"73,1,right,0"`
	Document          int    `fixed:"74,14,right,0"`
	CityDocument      int    `fixed:"88,8,right,0"`
	StateDocument     int    `fixed:"96,12,right,0"`
	Name              string `fixed:"108,75,left, "`
	AddressType       string `fixed:"183,3,left, "`
	Address           string `fixed:"186,50,left, "`
	AddressNumber     string `fixed:"236,10,left, "`
	AddressComplement string `fixed:"246,30,left, "`
	Neighborhood      string `fixed:"276,30,left, "`
	City              string `fixed:"306,50,left, "`
	State             string `fixed:"356,2,left, "`
	PostalCode        int    `fixed:"358,8,right,0"`
	Email             string `fixed:"366,75,left, "`
	Description       string `fixed:"441,500,left, "`
}

// footer represents the footer of the emission file.
type footer struct {
	RegType       int `fixed:"1,1,right,0"`
	TotalRecords  int `fixed:"2,7,right,0"`
	TotalAmount   int `fixed:"9,15,right,0"`
	TotalDiscount int `fixed:"24,15,right,0"`
}

// IssuerFile is a concrete implementation of the port.
type IssuerFile struct {
	filePath string
	pattern  string
	logger   port.Logger
	file     *os.File
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
	if err := i.openFile(emission); err != nil {
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

// openFile is a helper function to open the file for writing. 
func (i *IssuerFile) openFile(emission *domain.Emission) error {
	file_path := filepath.Join(i.filePath, i.replacePlaceholders(i.pattern, emission))
	file, err := os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	i.file = file
	return nil
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
	if _, err := i.file.Write(h); err != nil {
		return err
	}
	return nil
}

// writelines writes the emission lines to the file. 
func (i *IssuerFile) writeItems(emission *domain.Emission) error {
	emissionDate, _ := strconv.Atoi(emission.EmissionDate.Format("20060102"))
	for _, item := range emission.EmissionItems {
		it := i.getLine(emissionDate, &item)
		line, err := fixedwidth.Marshal(it)
		lineStr := string(line)
		re := regexp.MustCompile(`(?m) +$`)
		lineStr = re.ReplaceAllString(lineStr, "")
		if err != nil {
			return err
		}
		if _, err := i.file.WriteString(lineStr); err != nil {
			return err
		}
	}
	return nil
}

// getItem is a helper function to convert an EmissionItem to a line. 
func (i *IssuerFile) getLine(emissionDate int, item *domain.EmissionItem) line {
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
	return line{
		RegType:           1,
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
		Name:              item.Invoice.Customer.Name,
		AddressType:       " ",
		Address:           " ",
		AddressNumber:     " ",
		AddressComplement: " ",
		Neighborhood:      " ",
		City:              " ",
		State:             " ",
		PostalCode:        0,
		Email:             email,
		Description:       i.getDescription(item),
	}
}

// getDescription is a helper function to generate a description for the emission item.
func (i *IssuerFile) getDescription(item *domain.EmissionItem) string {
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
	if _, err := i.file.Write(f); err != nil {
		return err
	}
	return nil
}
