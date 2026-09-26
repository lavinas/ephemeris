package dto

import (
	"fmt"

	"github.com/klassmann/cpfcnpj"
	"github.com/nyaruka/phonenumbers"
)

var (
	validStatuses = map[string]bool{
		"realizada":            true,
		"cancelada_cobrar":     true,
		"cancelada_nao_cobrar": true,
	}
	validServices = map[string]bool{
		"aula/canto": true,
		"aula/piano": true,
	}
)

// ResponseBase represents the base structure for API responses, containing common fields for status and messages.
type ResponseBase struct {
	HttpCode int    `json:"http_code"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// NewResponseBase creates a new instance of ResponseBase with the provided HTTP code, status, and message.
func NewResponseBase(httpCode int, status, message string) ResponseBase {
	return ResponseBase{
		HttpCode: httpCode,
		Status:   status,
		Message:  message,
	}
}

// GetHTTPCode returns the HTTP status code of the response.
func (r ResponseBase) GetStatusCode() int {
	return r.HttpCode
}

// ValidateCellNumber checks if the provided phone number is a valid Brazilian cell phone number.
func ValidateCellNumber(phone string) (string, error) {
	num, err := phonenumbers.Parse(phone, "BR")
	if err != nil {
		return "", fmt.Errorf("invalid phone number format")
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number")
	}
	ptype := phonenumbers.GetNumberType(num)
	if ptype != phonenumbers.MOBILE && ptype != phonenumbers.UNKNOWN && ptype != phonenumbers.FIXED_LINE_OR_MOBILE {
		return "", fmt.Errorf("phone number is not a mobile number")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}

// ValidateDocument checks if the provided document is a valid CPF or CNPJ.
func ValidateDocument(document string, docType string) (string, error) {
	if docType == "cpf" || docType == "all" {
		cpf := cpfcnpj.NewCPF(document)
		if cpf.IsValid() {
			return cpf.String(), nil
		}
	}
	if docType == "cnpj" || docType == "all" {
		cnpj := cpfcnpj.NewCNPJ(document)
		if cnpj.IsValid() {
			return cnpj.String(), nil
		}
	}
	return "", fmt.Errorf("invalid document format")
}

