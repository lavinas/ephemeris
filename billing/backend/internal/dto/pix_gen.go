package dto

import (
	"strings"
	"errors"
	"billing/internal/port"
)


// PixRequest represents the request structure for generating a Pix payment payload.
type PixRequest struct {
	Key         string
	Description string
	Name        string
	City        string
	Amount      float64
	Txid        string
}

// Validate checks if the PixRequest has valid data.
func (r *PixRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if r.Key == "" {
		errs = append(errs, errors.New("pix key is required"))
	}
	if r.Description == "" {
		errs = append(errs, errors.New("description is required"))
	}
	if r.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if r.City == "" {
		errs = append(errs, errors.New("city is required"))
	}
	if r.Amount <= 0 {
		errs = append(errs, errors.New("amount must be greater than 0"))
	}
	if r.Txid == "" {
		errs = append(errs, errors.New("txid is required"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// Reset resets the PixRequest fields to their zero values.
func (r *PixRequest) Reset() {
	r.Key = ""
	r.Description = ""
	r.Name = ""
	r.City = ""
	r.Amount = 0
	r.Txid = ""
}


