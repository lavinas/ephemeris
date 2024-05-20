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
	// Read the CSV file into a slice of Record structs
	gocsv.SetCSVReader(u.setReader)
	gocsv.SetCSVWriter(u.setWriter)
	sessionsIn, err := u.readCSV(in.File)
	if err != nil {
		return err
	}
	sessionsOut := u.addSessions(sessionsIn)
	err = u.writeCSV(in.File, sessionsOut)
	if err != nil {
		return err
	}
	return nil
}

// setReader is a method that sets the reader
func (u *Usecase) setReader (r io.Reader) gocsv.CSVReader {
	reader := csv.NewReader(r)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	return reader
}

// setWriter is a method that sets the writer
func (u *Usecase) setWriter (out io.Writer) *gocsv.SafeCSVWriter {
	writer := csv.NewWriter(out)
	writer.Comma = ','
	return gocsv.NewSafeCSVWriter(writer)
}

// readCSV is a method that reads a csv
func (u *Usecase) readCSV(file string) ([]*dto.SessionCrud, error) {
	fileIn, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {
		return nil, u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	defer fileIn.Close()
	sessionsIn := []*dto.SessionCrud{}
	gocsv.UnmarshalFile(fileIn, &sessionsIn)
	return sessionsIn, nil
}

// writeCSV is a method that writes a csv
func (u *Usecase) writeCSV(file string, sessionsOut []*dto.SessionCSVOut) error {
	fileOut, err := os.OpenFile(file+".out.csv", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	defer fileOut.Close()
	gocsv.SetCSVWriter(u.setWriter)
	gocsv.MarshalFile(&sessionsOut, fileOut)
	return nil
}

// addSessions is a method that adds sessions
func (u *Usecase) addSessions(sessionsIn []*dto.SessionCrud) ([]*dto.SessionCSVOut) {
	sessionsOut := []*dto.SessionCSVOut{}
	for _, sessionIn := range sessionsIn {
		sessionIn.Action = "add"
		err := u.Add(sessionIn)
		result, message := u.getAddSessionResult(err)
		sessionsOut = append(sessionsOut, &dto.SessionCSVOut{
			ClientID: sessionIn.ClientID,
			ServiceID: sessionIn.ServiceID,
			At: sessionIn.At,
			Kind: sessionIn.Kind,
			Status: sessionIn.Status,
			Result: result,
			Message: message,
		})
	}
	return sessionsOut
}

// getAddSessionResult is a method that gets the result of adding a session
func (u *Usecase) getAddSessionResult(err error) (string, string) {
	result := ""
	message := ""
	if err != nil {
		result = "error"
		message = err.Error()
	} else {
		out := u.Out[0].(*dto.SessionCrud)
		result = "ok"
		message = fmt.Sprintf("Session %s added", out.ID)
	}
	return result, message
}