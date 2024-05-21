package dto

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/gocarina/gocsv"
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
		&PriceCrud{},
		&RecurrenceCrud{},
		&ServiceCrud{},
		&SessionCrud{},
	}
}

// Base represents the base dto
type Base struct {
}

// ReadCSV is a method that reads a csv
func (b *Base) ReadCSV(file string, dto interface{}) error {
	gocsv.SetCSVReader(b.setReader)
	fileIn, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer fileIn.Close()
	gocsv.UnmarshalFile(fileIn, dto)
	return nil
}

// setReader is a method that sets the reader
func (b *Base) setReader (r io.Reader) gocsv.CSVReader {
	reader := csv.NewReader(r)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	return reader
}
