package dto

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"billing/internal/domain"
	"billing/internal/port"
)

const (
	// requestLimit defines the maximum number of customer creation 
	// 	requests that can be processed in a single batch.
	requestLimit = 100
)

// CreateCustomerRequest represents the request payload for creating a new customer.
type CreateCustomerRequest struct {
	Items []CreateCustomerRequestItem `json:"items" validate:"required,dive"`
}

// CreateCustomerRequestItem represents an individual customer creation request item.
type CreateCustomerRequestItem struct {
	Name     string `json:"name" validate:"required"`
	Nickname string `json:"nickname" validate:"required"`
	Document string `json:"document" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Whatsapp string `json:"whatsapp" validate:"required"`
}

// CreateCustomerResponse represents the response payload after creating a new customer.
type CreateCustomerResponse struct {
	ResponseBase
}

// NewCreateCustomerResponse creates a new instance of CreateCustomerResponse 
func NewCreateCustomerResponse(httpCode int16, status, message string) CreateCustomerResponse {
	return CreateCustomerResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the CreateCustomerRequest has all required fields and valid data.
func (r *CreateCustomerRequest) Validate(repo port.Repository) error {
	if len(r.Items) == 0 {
		return errors.New("no customer data provided")
	}
	if len(r.Items) > requestLimit {
		return fmt.Errorf("too many customer creation requests: maximum is %d", requestLimit)
	}
	errs := r.validateItems(repo)
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateItems validate items and duplicated items in the request
func (r *CreateCustomerRequest) validateItems(repo port.Repository) []error {
	documents := make(map[string]bool, len(r.Items))
	nicknames := make(map[string]bool, len(r.Items))
	errs := make([]error, 0)
	for i, item := range r.Items {
		if err := item.Validate(repo); err != nil {
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

// Validate checks if the CreateCustomerRequestItem has all required fields and valid data.
func (r *CreateCustomerRequestItem) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDocument(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsapp(repo); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain converts the CreateCustomerRequest to a slice of domain.Customer entities.
func (r *CreateCustomerRequest) GetDomain() []*domain.Customer {
	customers := make([]*domain.Customer, len(r.Items))
	for i, item := range r.Items {
		customers[i] = item.GetDomain()
	}
	return customers
}

// GetDomain converts the CreateCustomerRequestItem to a domain.Customer entity.
func (r *CreateCustomerRequestItem) GetDomain() *domain.Customer {
	return domain.NewCustomer(r.Name, r.Nickname, r.Document, r.Email, r.Whatsapp)
}

// validateName checks if the provided name is valid.
func (r *CreateCustomerRequestItem) validateName() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// validateNickname checks if the provided nickname is valid and not already in use.
func (r *CreateCustomerRequestItem) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		return errors.New("nickname is required")
	}
	existingCustomer, err := repo.GetCustomerByNickname(r.Nickname)
	if err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	}
	if existingCustomer != nil {
		return fmt.Errorf("nickname is already in use")
	}
	return nil
}

// validateDocument checks if the provided document is valid and not already in use.
func (r *CreateCustomerRequestItem) validateDocument(repo port.Repository) error {
	if r.Document == "" {
		return nil
	}
	existingCustomer, err := repo.GetCustomerByDocument(r.Document)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}
	if existingCustomer != nil {
		return fmt.Errorf("document is already in use")
	}
	return nil
}

// validateEmail checks if the provided email is valid and not already in use.
func (r *CreateCustomerRequestItem) validateEmail() error {
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
func (r *CreateCustomerRequestItem) validateWhatsapp(repo port.Repository) error {
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
	return nil
}