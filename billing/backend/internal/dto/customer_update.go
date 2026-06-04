package dto

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"billing/internal/domain"
	"billing/internal/port"

	"github.com/klassmann/cpfcnpj"
)

// CustomerUpdateRequest represents the request payload for updating an existing customer.
type CustomerUpdateRequest struct {
	Nickname string           `json:"nickname" validate:"required"`
	Name     string           `json:"name" validate:"required"`
	Document string           `json:"document" validate:"required"`
	Email    string           `json:"email" validate:"required,email"`
	Whatsapp string           `json:"whatsapp" validate:"required"`
	Status   *int             `json:"status" validate:"required"`
	customer *domain.Customer `json:"-" validate:"-"`
}

// CustomerUpdateResponse represents the response payload after updating an existing customer.
type CustomerUpdateResponse struct {
	ResponseBase
}

// NewCustomerUpdateResponse creates a new instance of CustomerUpdateResponse
func NewCustomerUpdateResponse(httpCode int, status, message string) CustomerUpdateResponse {
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
	if err := r.validateStatus(); err != nil {
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
	r.customer = customer
	return nil
}

// validateDocument checks if the provided document is valid and not used by another customer.
func (r *CustomerUpdateRequest) validateDocument(repo port.Repository) error {
	if r.Document == "" {
		return nil
	}
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	exist, err := repo.FindCustomers(0, 0, 0, nil, nil, &r.Document, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}

	if len(exist) > 0 && r.customer != nil && exist[0].ID != r.customer.ID {
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

// validateStatus checks if the provided status is either 0 (inactive) or 1 (active).
func (r *CustomerUpdateRequest) validateStatus() error {
	if r.Status != nil && *r.Status != 0 && *r.Status != 1 {
		return fmt.Errorf("status must be either 0 (inactive) or 1 (active)")
	}
	return nil
}

// validateAtLesatOneField checks if at least one of the fields is provided for update.
func (r *CustomerUpdateRequest) validateAtLeastOneField() error {
	if r.Name == "" && r.Document == "" && r.Email == "" && r.Whatsapp == "" && r.Status == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// GetModel constructs a map of the fields to be updated based on the non-empty fields
func (r *CustomerUpdateRequest) GetDomain() interface{} {
	if r.customer == nil {
		return nil
	}
	if r.Name != "" {
		r.customer.Name = r.Name
	}
	if r.Document != "" {
		r.customer.Document = &r.Document
	}
	if r.Email != "" {
		r.customer.Email = &r.Email
	}
	if r.Whatsapp != "" {
		r.customer.Whatsapp = &r.Whatsapp
	}
	if r.Status != nil {
		r.customer.Status = *r.Status
	}
	return r.customer
}

// Reset resets the fields of the CustomerUpdateRequest to their zero values.
func (r *CustomerUpdateRequest) Reset() {
	r.Nickname = ""
	r.Name = ""
	r.Document = ""
	r.Email = ""
	r.Whatsapp = ""
	r.Status = nil
	r.customer = nil
}
