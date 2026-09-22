package domain

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/klassmann/cpfcnpj"
	"github.com/nyaruka/phonenumbers"
	"signup/internal/port"
)

// DomainBase represents the base structure for all domain models.
type DomainBase struct {
	Repo port.Repository
}

// ValidatePhoneNumber checks if the provided phone number is valid and formats it to E.164 standard.
func (b *DomainBase) ValidateCellNumber(phone string) (string, error) {
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
func (b *DomainBase) ValidateEmail(email string) error {
	if email != strings.ToLower(email) || strings.Contains(email, " ") {
		return fmt.Errorf("email must be lowercase and must not contain spaces")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateCpfCnpj checks if the provided document is a valid CPF or CNPJ.
func (b *DomainBase) ValidateCpfCnpj(document string) (string, error) {
	cpf := cpfcnpj.NewCPF(document)
	if cpf.IsValid() {
		return cpf.String(), nil
	}
	cnpj := cpfcnpj.NewCNPJ(document)
	if cnpj.IsValid() {
		return cnpj.String(), nil
	}
	return "", fmt.Errorf("invalid document format")
}

// ValidateNickname checks if the provided nickname is valid and not already in use.
func (b *DomainBase) ValidateNickname(nickname string) error {
	if nickname != strings.ToLower(nickname) || strings.Contains(nickname, " ") {
		return fmt.Errorf("nickname must be lowercase and must not contain spaces")
	}
	return nil
}

// ValidatePassword checks if the provided password is valid and not already in use.
func (b *DomainBase) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !strings.ContainsAny(password, "0123456789") {
		return fmt.Errorf("password must contain at least one number")
	}
	if !strings.ContainsAny(password, "!@#$%^&*") {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
