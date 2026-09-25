package dto

import (
	"fmt"
	"github.com/nyaruka/phonenumbers"
	"net/mail"
	"strings"

	"signup/internal/port"

	"github.com/klassmann/cpfcnpj"
	"golang.org/x/crypto/bcrypt"
)

// RequestBase embeds Repository.
type RequestBase struct {
	Repo port.Repository
}

// ResponseBase represents the base structure for API responses, containing common fields for status and messages.
type ResponseBase struct {
	HttpCode int    `json:"http_code"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// NewRequestBase creates a new instance of RequestBase with the provided repository.
func NewRequestBase(repo port.Repository) RequestBase {
	return RequestBase{
		Repo: repo,
	}
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

// ValidatePhoneNumber checks if the provided phone number is valid and formats it to E.164 standard.
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

// ValidateEmail checks if the provided email is valid and not already in use.
func ValidateEmail(email string) error {
	if email != strings.ToLower(email) || strings.Contains(email, " ") {
		return fmt.Errorf("email must be lowercase and must not contain spaces")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// HashPassword hashes the provided password.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// ValidateDocument checks if the provided document is a valid CPF or CNPJ.
func ValidateDocument(document string, docType string) (string, error) {
// validateCpfCnpj checks if the provided document is a valid CPF or CNPJ.
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
