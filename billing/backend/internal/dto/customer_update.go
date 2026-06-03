package dto

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"net/mail"

	"billing/internal/port"
	"github.com/klassmann/cpfcnpj"
)

// CustomerUpdateRequest represents the request payload for updating an existing customer.
type CustomerUpdateRequest struct {
	Nickname string `json:"nickname" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Document string `json:"document" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Whatsapp string `json:"whatsapp" validate:"required"`
	ID       int64  `json:"-" validate:"-"`
}

// CustomerUpdateResponse represents the response payload after updating an existing customer.
type CustomerUpdateResponse struct {
	ResponseBase
}

// NewCustomerUpdateResponse creates a new instance of CustomerUpdateResponse
func NewCustomerUpdateResponse(httpCode int16, status, message string) CustomerUpdateResponse {
	return CustomerUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the CustomerUpdateRequest has all required fields and valid data.
func (r *CustomerUpdateRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	} 
	if err := r.validateDocument(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsapp(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAtLeastOneField(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}


// validateNickname checks if the provided nickname is not already used by another customer.
func (r *CustomerUpdateRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
    customer, err := repo.GetCustomer(r.Nickname)
	if err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	}
	if customer == nil {
		return fmt.Errorf("customer with nickname '%s' does not exist", r.Nickname)
	}
	return nil
}


// validateDocument checks if the provided document is valid and not already used by another customer.
func (r *CustomerUpdateRequest) validateDocument(repo port.Repository) error {
	if r.Document == "" {
		return nil
	}
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	existingCustomer, err := repo.FindCustomers(0, 0, 0, nil, nil, &r.Document, -1, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}
	if len(existingCustomer) > 0 && existingCustomer[0].ID != r.ID {
		return fmt.Errorf("document '%s' is already in use by another customer", r.Document)
	}
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is in a valid format and formats it.
func (r *CustomerUpdateRequest) validateWhatsapp() error {
	if r.Whatsapp == "" {
		return nil
	}
	matched, err := regexp.MatchString(`^\(?\d{2}\)?\s?\d{4,5}-?\d{4}$`, r.Whatsapp)
	if err != nil {
		return fmt.Errorf("failed to validate whatsapp: %v", err)
	}
	if !matched {
		return fmt.Errorf("whatsapp '%s' is not in a valid format", r.Whatsapp)
	}
	return nil
}

// validateCpfCnpj checks if the provided document is a valid CPF or CNPJ and formats it.
func (r *CustomerUpdateRequest) validateCpfCnpj() error {
	if r.Document == "" {
		return nil
	}
	cpf := cpfcnpj.NewCPF(r.Document)
	if cpf.IsValid() {
		r.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(r.Document)
	if cnpj.IsValid() {
		r.Document = cnpj.String()
		return nil
	}
	return fmt.Errorf("document '%s' is not a valid CPF or CNPJ", r.Document)
}

// validateEmail checks if the provided email is in a valid format.
func (r *CustomerUpdateRequest) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return fmt.Errorf("email '%s' is not in a valid format", r.Email)
	}
	return nil
}

// validateAtLesatOneField checks if at least one of the fields is provided for update.
func (r *CustomerUpdateRequest) validateAtLeastOneField() error {
	if r.Name == "" && r.Document == "" && r.Email == "" && r.Whatsapp == "" {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// GetModel constructs a map of the fields to be updated based on the non-empty fields in the request.
func (r *CustomerUpdateRequest) GetDomain() interface{} {
	ret := make(map[string]interface{})
	if r.Name != "" {
		ret["name"] = r.Name
	}
	if r.Document != "" {
		ret["document"] = r.Document
	}
	if r.Email != "" {
		ret["email"] = r.Email
	}
	if r.Whatsapp != "" {
		ret["whatsapp"] = r.Whatsapp
	}
	return ret
}
