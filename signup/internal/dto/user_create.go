package dto

import (
	"errors"
	"fmt"
	"strings"

	"signup/internal/domain"
	"signup/internal/port"
)

// UserCreateRequest represents the request payload for creating a new user.
type UserCreateRequest struct {
	RequestBase `json:"-"`
	Vendor      string `json:"vendor" validate:"required"`
	VendorID    int64  `json:"_"`
	Name        string `json:"name" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Whatsapp    string `json:"whatsapp" validate:"required"`
}

// NewUserCreateRequest creates a new instance of UserCreateRequest.
func NewUserCreateRequest(repo port.Repository) *UserCreateRequest {
	return &UserCreateRequest{RequestBase: RequestBase{Repo: repo}}
}

// UserCreateResponse represents the response payload after creating a new user.
type UserCreateResponse struct {
	ResponseBase
}

// NewUserCreateResponse creates a new instance of UserCreateResponse.
func NewUserCreateResponse(httpCode int, status, message string) UserCreateResponse {
	return UserCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the UserCreateRequest has all required fields and valid data.
func (r *UserCreateRequest) Validate() error {
	errs := []error{}
	if err := r.validateVendor(); err != nil {
		errs = append(errs, fmt.Errorf("vendor is required"))
	}
	if err := r.validateName(); err != nil {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if err := r.validateUser(); err != nil {
		errs = append(errs, fmt.Errorf("username is required"))
	}
	if err := r.validatePassword(); err != nil {
		errs = append(errs, fmt.Errorf("password is required"))
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, fmt.Errorf("email is required"))
	}
	if err := r.validateWhatsapp(); err != nil {
		errs = append(errs, fmt.Errorf("whatsapp is required"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain returns the domain model of the user.
func (r *UserCreateRequest) GetDomain() port.Domain {
	var email, whatsapp *string
	if r.Email != "" {
		email = &r.Email
	}
	if r.Whatsapp != "" {
		whatsapp = &r.Whatsapp
	}
	return domain.NewUser(r.Repo, r.VendorID, r.Name, r.Username, r.Password, email, whatsapp)
}

// validateVendor checks if the provided vendor is valid.
func (r *UserCreateRequest) validateVendor() error {
	if r.Vendor == "" {
		return errors.New("vendor is required")
	}
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByNickname(r.Vendor); err != nil || !ok {
		return errors.New("invalid vendor")
	}
	r.VendorID = vendor.ID
	return nil
}

// validateName checks if the provided name is valid.
func (r *UserCreateRequest) validateName() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// validateUser checks if the provided user data is valid.
func (r *UserCreateRequest) validateUser() error {
	if r.Username == "" {
		return errors.New("username is required")
	}
	// verificar se username tem apenas caracteres minúsculos e não contém espaços
	if r.Username != strings.ToLower(r.Username) || strings.Contains(r.Username, " ") {
		return errors.New("username must be lowercase and must not contain spaces")
	}
	// if it contains spaces, the previous condition will catch it
	user := domain.StartUser(r.Repo)
	if ok, err := user.GetByUsername(r.VendorID, r.Username); err != nil {
		return fmt.Errorf("error checking username: %w", err)
	} else if ok {
		return fmt.Errorf("username already exists")
	}
	return nil
}

// validatePassword checks if the provided password is valid.
func (r *UserCreateRequest) validatePassword() error {
	if r.Password == "" {
		return errors.New("password is required")
	}
	if strings.Contains(r.Password, " ") {
		return errors.New("password must not contain spaces")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	var err error
	r.Password, err = HashPassword(r.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}
	return nil
}

// validateEmail checks if the provided email is valid.
func (r *UserCreateRequest) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}
	// check if email already exists for the vendor
	user := domain.StartUser(r.Repo)
	if ok, err := user.GetByEmail(r.VendorID, r.Email); err != nil {
		return fmt.Errorf("error checking email: %w", err)
	} else if ok {
		return fmt.Errorf("email already exists")
	}
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is valid and not already in use.
func (r *UserCreateRequest) validateWhatsapp() error {
	if r.Whatsapp == "" {
		return nil
	}
	num, err := ValidateCellNumber(r.Whatsapp)
	if err == nil {
		r.Whatsapp = num
		return nil
	}
	num, err = ValidateCellNumber(fmt.Sprintf("+%s", r.Whatsapp))
	if err == nil {
		r.Whatsapp = num
		return nil
	}
	return fmt.Errorf("invalid WhatsApp number format")
}
