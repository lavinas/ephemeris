package usecase

import (
	"encoding/csv"
	"os"
	"fmt"

	"github.com/lavinas/ephemeris/internal/dto"
	"github.com/lavinas/ephemeris/pkg"
)

// GetCSV is a method that gets a csv
func (u *Usecase) SessionCSV(dtoIn interface{}) error {
	in := dtoIn.(*dto.SessionCSV)
	if err := in.Validate(u.Repo); err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	file, err := os.Open(in.File)
	if err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	defer file.Close()
	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		return u.error(pkg.ErrPrefBadRequest, err.Error())
	}
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}
