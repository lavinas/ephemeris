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

const (
	// requestLimit defines the maximum number of customer creation
	// 	requests that can be processed in a single batch.
	requestLimit = 100
)

// CustomerCreateRequest represents the request payload for creating a new customer.
type CustomerCreateRequest struct {
	Vendor   string                       `json:"vendor" validate:"required"`
	VendorID int64                        `json:"-" validate:"-"`
	Items    []*CustomerCreateRequestItem `json:"items" validate:"required,dive"`
}

// CustomerCreateRequestItem represents an individual customer creation request item.
type CustomerCreateRequestItem struct {
	Name     string `json:"name" validate:"required"`
	Nickname string `json:"nickname" validate:"required"`
	Document string `json:"document" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Whatsapp string `json:"whatsapp" validate:"required"`
}

// CustomerCreateResponse represents the response payload after creating a new customer.
type CustomerCreateResponse struct {
	ResponseBase
}

// NewCustomerCreateResponse creates a new instance of CustomerCreateResponse
func NewCustomerCreateResponse(httpCode int, status, message string) CustomerCreateResponse {
	return CustomerCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the CustomerCreateRequest has all required fields and valid data.
func (r *CustomerCreateRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateVendor(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateItems(repo); err != nil {
		errs = append(errs, err...)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateVendor checks if the provided vendor is valid and exists in the system.
func (r *CustomerCreateRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return errors.New("vendor is required")
	}
	vendor, err := repo.GetVendor(r.Vendor)
	if err != nil {
		return fmt.Errorf("failed to validate vendor: %v", err)
	}
	if vendor == nil {
		return fmt.Errorf("vendor '%s' does not exist", r.Vendor)
	}
	r.VendorID = vendor.ID
	return nil
}

// validateItems validate items and duplicated items in the request
func (r *CustomerCreateRequest) validateItems(repo port.Repository) []error {
	if len(r.Items) == 0 {
		return []error{errors.New("at least one item is required")}
	}
	if len(r.Items) > requestLimit {
		return []error{fmt.Errorf("number of items exceeds the limit of %d", requestLimit)}
	}
	documents := make(map[string]bool, len(r.Items))
	nicknames := make(map[string]bool, len(r.Items))
	errs := make([]error, 0)
	for i, item := range r.Items {
		if err := item.Validate(repo, r.VendorID); err != nil {
			errs = append(errs, fmt.Errorf("item %d: %v", i, err))
		}
		if item.Document != "" && documents[item.Document] {
			errs = append(errs, fmt.Errorf("item %d: duplicate document '%s'", i, item.Document))
		}
		if item.Nickname != "" && nicknames[item.Nickname] {
			errs = append(errs, fmt.Errorf("item %d: duplicate nickname '%s'", i, item.Nickname))
		}
		documents[item.Document] = true
		nicknames[item.Nickname] = true
	}
	return errs
}

// Validate checks if the CustomerCreateRequestItem has all required fields and valid data.
func (r *CustomerCreateRequestItem) Validate(repo port.Repository, vendorID int64) error {
	errs := make([]error, 0)
	if err := r.validateName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(repo, vendorID); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDocument(repo, vendorID); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsapp(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain converts the CustomerCreateRequest to a slice of domain.Customer entities.
func (r *CustomerCreateRequest) GetDomain() interface{} {
	customers := make([]domain.Customer, len(r.Items))
	for i, item := range r.Items {
		customers[i] = *item.GetDomain(r.VendorID)
	}
	return &customers
}

// GetDomain converts the CustomerCreateRequestItem to a domain.Customer entity.
func (r *CustomerCreateRequestItem) GetDomain(vendorID int64) *domain.Customer {
	var document, email, whatsapp *string
	if r.Document != "" {
		document = &r.Document
	}
	if r.Email != "" {
		email = &r.Email
	}
	if r.Whatsapp != "" {
		whatsapp = &r.Whatsapp
	}
	return domain.NewCustomer(vendorID, &r.Name, &r.Nickname, document, email, whatsapp)
}

// validateName checks if the provided name is valid.
func (r *CustomerCreateRequestItem) validateName() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// validateNickname checks if the provided nickname is valid and not already in use.
func (r *CustomerCreateRequestItem) validateNickname(repo port.Repository, vendorID int64) error {
	if r.Nickname == "" {
		return errors.New("nickname is required")
	}
	existingCustomer, err := repo.FindCustomers(0, 0, vendorID, nil, &r.Nickname,
		nil, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	}
	if len(existingCustomer) > 0 {
		return fmt.Errorf("nickname is already in use")
	}
	return nil
}

// validateDocument checks if the provided document is valid and not already in use.
func (r *CustomerCreateRequestItem) validateDocument(repo port.Repository, vendorID int64) error {
	if r.Document == "" {
		return nil
	}
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	existingCustomer, err := repo.FindCustomers(0, 0, vendorID, nil, nil, &r.Document, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}
	if len(existingCustomer) > 0 {
		return fmt.Errorf("document is already in use")
	}
	return nil
}

// validateCpfCnpj checks if the provided document is a valid CPF or CNPJ.
func (r *CustomerCreateRequestItem) validateCpfCnpj() error {
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
	return fmt.Errorf("invalid document format")
}

// validateEmail checks if the provided email is valid and not already in use.
func (r *CustomerCreateRequestItem) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is valid and not already in use.
func (r *CustomerCreateRequestItem) validateWhatsapp() error {
	if r.Whatsapp == "" {
		return nil
	}
	phoneRegex := `^\(?([1-9]{2})\)?\s?(9[1-9][0-9]{3})-?([0-9]{4})$`
	matched, err := regexp.MatchString(phoneRegex, r.Whatsapp)
	if err != nil {
		return fmt.Errorf("failed to validate whatsapp: %v", err)
	}
	if !matched {
		return fmt.Errorf("invalid whatsapp format")
	}
	// formating the WhatsApp number to only digits to extract the area code and the number
	n := regexp.MustCompile(`\D`).ReplaceAllString(r.Whatsapp, "")
	r.Whatsapp = fmt.Sprintf("(%s) %s-%s", n[0:2], n[2:7], n[7:11])
	return nil
}

// Reset resets the fields of the CustomerCreateRequest to their zero values.
func (r *CustomerCreateRequest) Reset() {
	r.Vendor = ""
	r.VendorID = 0
	r.Items = nil
}
