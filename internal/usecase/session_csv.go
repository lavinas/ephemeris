package usecase

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// GetCSV is a method that gets a csv
func (u *Usecase) SessionCSV(dtoIn interface{}) error {
	in := dtoIn.(*dto.SessionCSV)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	fileIn, err := os.OpenFile(in.File, os.O_RDONLY, 0644)
	if err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	defer fileIn.Close()
	// Read the CSV file into a slice of Record structs
	gocsv.SetCSVReader(func(r io.Reader) gocsv.CSVReader {
        reader := csv.NewReader(r)
        reader.Comma = ','
        reader.LazyQuotes = true
        reader.FieldsPerRecord = -1
        return reader
    })
	fileOut, err := os.OpenFile(in.File+".out.csv", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = ','
		return gocsv.NewSafeCSVWriter(writer)
	})
	defer fileOut.Close()
	sessionsIn := []*dto.SessionCrud{}
	gocsv.UnmarshalFile(fileIn, &sessionsIn)
	sessionsOut := []*dto.SessionCSVOut{}
	for _, sessionIn := range sessionsIn {
		sessionIn.Action = "add"
		err := u.Add(sessionIn)
		result := "added"
		id := ""
		if err != nil {
			result = fmt.Sprintf("error: %s", err.Error())
		} else {
			out := u.Out[0].(*dto.SessionCrud)
			id = out.ID
		}
		sessionsOut = append(sessionsOut, &dto.SessionCSVOut{
			ClientID: sessionIn.ClientID,
			ServiceID: sessionIn.ServiceID,
			At: sessionIn.At,
			Kind: sessionIn.Kind,
			Status: sessionIn.Status,
			Result: result,
			ID: id,
		})
	}
	gocsv.MarshalFile(&sessionsOut, fileOut)
	return nil
}
