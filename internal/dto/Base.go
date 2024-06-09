package dto

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/lavinas/ephemeris/internal/port"
	"github.com/lavinas/ephemeris/pkg"
)

func All() []interface{} {
	return []interface{}{
		&AgendaCrud{},
		&AgendaMake{},
		&ClientCrud{},
		&ContractCrud{},
		&InvoiceCrud{},
		&InvoiceItemCrud{},
		&PackageCrud{},
		&PackageAppend{},
		&RecurrenceCrud{},
		&ServiceCrud{},
		&SessionCrud{},
		&SessionTie{},
	}
}

// Base represents the base dto
type Base struct {
}

// ReadCSV is a method that reads a csv
func (b *Base) ReadCSV(dto interface{}, file string) error {
	gocsv.SetCSVReader(b.setReader)
	fileIn, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer fileIn.Close()
	gocsv.UnmarshalFile(fileIn, dto)
	return nil
}

// Getinstructions is a method that returns the instructions of the dto for given domain
func (b *Base) getInstructions(s port.DTOIn, domain port.Domain) (port.Domain, []interface{}, error) {
	cmd, err := pkg.NewCommands().Transpose(s)
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) > 0 {
		domain := s.GetDomain()[0]
		return domain, cmd, nil
	}
	return domain, cmd, nil
}

// setReader is a method that sets the reader
func (b *Base) setReader(r io.Reader) gocsv.CSVReader {
	reader := csv.NewReader(r)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	return reader
}
