package dto

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"signup/internal/domain"
	"signup/internal/port"

	"github.com/klassmann/cpfcnpj"
)

// CustomerUpdateRequest represents the request payload for updating an existing customer.
type CustomerUpdateRequest struct {
	RequestBase `json:"-" validate:"-"`
	Vendor      string           `json:"vendor" validate:"required"`
	vendorID    int64            `json:"-" validate:"-"`
	Nickname    string           `json:"nickname" validate:"required"`
	Name        *string          `json:"name" validate:"required"`
	Document    *string          `json:"document" validate:"required"`
	Email       *string          `json:"email" validate:"required,email"`
	Whatsapp    *string          `json:"whatsapp" validate:"required"`
	Status      *int             `json:"status" validate:"required"`
	customer    *domain.Customer `json:"-" validate:"-"`
}

// NewCustomerUpdateRequest creates a new instance of CustomerUpdateRequest
func NewCustomerUpdateRequest(repo port.Repository) *CustomerUpdateRequest {
	return &CustomerUpdateRequest{
		RequestBase: NewRequestBase(repo),
	}
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
func (r *CustomerUpdateRequest) Validate() error {
	errs := make([]error, 0)
	if err := r.validateVendor(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDocument(); err != nil {
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

// GetDomain returns the domain model of the customer.
func (r *CustomerUpdateRequest) GetDomain() (port.Domain, error) {
	if r.customer == nil {
		return nil, errors.New("customer not found")
	}
	if r.Name != nil {
		r.customer.Name = *r.Name
	}
	if r.Document != nil {
		r.customer.Document = r.Document
	}
	if r.Email != nil {
		r.customer.Email = r.Email
	}
	if r.Whatsapp != nil {
		r.customer.Whatsapp = r.Whatsapp
	}
	if r.Status != nil {
		r.customer.Status = r.Status
	}
	return r.customer, nil
}

// GetOutDTO converts the CustomerCreateRequest to a CustomerCreateResponse DTO.
func (r *CustomerUpdateRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	return &CustomerUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// GetPageParams returns the pagination parameters.
func (r *CustomerUpdateRequest) GetPageParams() (int, int) {
	return 1, 1
}

// ValidateVendor checks if the provided vendor is valid and sets the vendorID.
func (r *CustomerUpdateRequest) validateVendor() error {
	if r.Vendor == "" {
		return fmt.Errorf("vendor is required")
	}
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByNickname(r.Vendor); err != nil {
		return fmt.Errorf("failed to validate vendor: %v", err)
	} else if !ok {
		return fmt.Errorf("vendor '%s' does not exist", r.Vendor)
	}
	r.vendorID = vendor.ID
	return nil
}

// validateNickname checks if the provided nickname is not already used by another customer.
func (r *CustomerUpdateRequest) validateNickname() error {
	if r.vendorID == 0 {
		return nil // Vendor validation will catch this error
	}
	if r.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	customer := domain.StartCustomer(r.Repo)
	if ok, err := customer.GetByNickname(r.vendorID, r.Nickname); err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	} else if !ok {
		return fmt.Errorf("customer with nickname '%s' does not exist", r.Nickname)
	}
	r.customer = customer
	return nil
}

// validateName checks if the provided name is not empty.
func (r *CustomerUpdateRequest) validateName() error {
	if r.Name == nil {
		return nil
	}
	if *r.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

// validateDocument checks if the provided document is valid and not used by another customer.
func (r *CustomerUpdateRequest) validateDocument() error {
	if r.Document == nil {
		return nil
	}
	if *r.Document == "" {
		return fmt.Errorf("document cannot be empty")
	}
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	customer := domain.StartCustomer(r.Repo)
	if ok, err := customer.GetByDocument(r.vendorID, *r.Document); err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	} else if !ok {
		return nil
	}
	if customer.ID == r.customer.ID {
		return nil
	}
	return fmt.Errorf("document '%s' is already in use by another customer", *r.Document)
}

// validateWhatsapp checks if the provided WhatsApp number is in a valid format and formats it.
func (r *CustomerUpdateRequest) validateWhatsapp() error {
	if r.Whatsapp == nil {
		return nil
	}
	if *r.Whatsapp == "" {
		return fmt.Errorf("whatsapp cannot be empty")
	}
	num, err := ValidateCellNumber(*r.Whatsapp)
	if err == nil {
		r.Whatsapp = &num
		return nil
	}
	num, err = ValidateCellNumber(fmt.Sprintf("+%s", *r.Whatsapp))
	if err == nil {
		r.Whatsapp = &num
		return nil
	}
	return fmt.Errorf("invalid WhatsApp number format")
}

// validateCpfCnpj checks if the provided document is a valid CPF or CNPJ and formats it.
func (r *CustomerUpdateRequest) validateCpfCnpj() error {
	if r.Document == nil {
		return nil
	}
	cpf := cpfcnpj.NewCPF(*r.Document)
	if cpf.IsValid() {
		*r.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(*r.Document)
	if cnpj.IsValid() {
		*r.Document = cnpj.String()
		return nil
	}
	return fmt.Errorf("document '%s' is not a valid CPF or CNPJ", *r.Document)
}

// validateEmail checks if the provided email is in a valid format.
func (r *CustomerUpdateRequest) validateEmail() error {
	if r.Email == nil {
		return nil
	}
	if *r.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	_, err := mail.ParseAddress(*r.Email)
	if err != nil {
		return fmt.Errorf("email '%s' is not in a valid format", *r.Email)
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
	if r.Name == nil && r.Document == nil && r.Email == nil && r.Whatsapp == nil &&
		r.Status == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// Reset resets the fields of the CustomerUpdateRequest to their zero values.
func (r *CustomerUpdateRequest) Reset() {
	r.Vendor = ""
	r.vendorID = 0
	r.Nickname = ""
	r.Name = nil
	r.Document = nil
	r.Email = nil
	r.Whatsapp = nil
	r.Status = nil
	r.customer = nil
}
