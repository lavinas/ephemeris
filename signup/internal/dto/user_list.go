package dto

import (
	"errors"
	"fmt"
	"strings"

	"signup/internal/domain"
	"signup/internal/port"
)

// UserListRequest represents the request payload for listing users.
type UserListRequest struct {
	RequestBase `json:"-" validate:"-"`
	Page        int     `json:"page" validate:"required,gt=0"`
	PageSize    int     `json:"page_size" validate:"required,gt=0"`
	Vendor      string  `json:"vendor" validate:"required"`
	VendorID    int64   `json:"-" validate:"-"`
	Name        *string `json:"name,omitempty"`
	Username    *string `json:"username,omitempty"`
	Status      *int    `json:"status,omitempty"`
	Email       *string `json:"email,omitempty"`
	Whatsapp    *string `json:"whatsapp,omitempty"`
}

// NewUserListRequest creates a new UserListRequest.
func NewUserListRequest(repo port.Repository) *UserListRequest {
	return &UserListRequest{
		RequestBase: NewRequestBase(repo),
	}
}

// UserListResponse represents the response payload for listing users.
type UserListResponse struct {
	ResponseBase
	Users      []UserListItem `json:"users"`
	TotalCount int            `json:"total_count"`
}

// UserListItem represents a single user in the list response.
type UserListItem struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Whatsapp string `json:"whatsapp"`
	Status   int    `json:"status"`
}

// NewUserListResponse creates a new instance of CustomerListResponse with the provided
// HTTP code, status, message, and customers.
func NewUserListResponse(httpCode int, status, message string,
	customers []UserListItem) UserListResponse {
	return UserListResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		Users:        customers,
		TotalCount:   len(customers),
	}
}

// NewUserListItem creates a new UserListItem with the given user details.
func NewUserListItem(id int64, username, name, email, whatsapp string, status int) UserListItem {
	return UserListItem{
		ID:       id,
		Username: username,
		Name:     name,
		Email:    email,
		Whatsapp: whatsapp,
		Status:   status,
	}
}

// Validate checks if the UserListRequest has valid fields.
func (r *UserListRequest) Validate() error {
	errs := []error{}
	if err := r.validatePage(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePageSize(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateVendor(); err != nil {
		errs = append(errs, err)
	}
	if r.Status != nil {
		if *r.Status < 0 || *r.Status > 1 {
			errs = append(errs, fmt.Errorf("status must be 0 or 1"))
		}
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain converts the UserListRequest to a domain.User entity.
func (r *UserListRequest) GetDomain() (port.Domain, error) {
	user := domain.StartUser(r.Repo)
	user.VendorID = r.VendorID
	if r.Name != nil {
		user.Name = *r.Name
	}
	if r.Username != nil {
		user.Username = *r.Username
	}
	user.Email = r.Email
	user.Whatsapp = r.Whatsapp
	user.Status = r.Status
	return user, nil
}

// GetOutDTO constructs an output DTO for the user list request.
func (r *UserListRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	users := make([]UserListItem, 0, 10)
	if list, ok := data.([]port.Domain); ok {
		for _, d := range list {
			if u, ok := d.(*domain.User); ok {
				email := "-"
				if u.Email != nil {
					email = *u.Email
				}
				whatsapp := "-"
				if u.Whatsapp != nil {
					whatsapp = *u.Whatsapp
				}
				statusVal := 0
				if u.Status != nil {
					statusVal = *u.Status
				}
				users = append(users, NewUserListItem(u.ID, u.Username, u.Name, email, whatsapp, statusVal))
			}
		}
		return NewUserListResponse(httpCode, status, message, users)
	}
	return NewUserListResponse(httpCode, status, message, []UserListItem{})
}

// GetPageParams returns the pagination parameters.
// return (page, pageSize)
func (r *UserListRequest) GetPageParams() (int, int) {
	return r.Page, r.PageSize
}

// validatePage checks if the provided page number is valid.
func (r *UserListRequest) validatePage() error {
	if r.Page <= 0 {
		return fmt.Errorf("page must be greater than 0")
	}
	return nil
}

// validatePageSize checks if the provided page size is valid.
func (r *UserListRequest) validatePageSize() error {
	if r.PageSize <= 0 {
		return fmt.Errorf("page_size must be greater than 0")
	}
	return nil
}

// validateVendor checks if the provided vendor is valid.
func (r *UserListRequest) validateVendor() error {
	if r.Vendor == "" {
		return fmt.Errorf("vendor is required")
	}
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByNickname(r.Vendor); err != nil {
		return fmt.Errorf("error fetching vendor by nickname: %w", err)
	} else if !ok {
		return fmt.Errorf("vendor is not valid")
	}
	r.VendorID = vendor.ID
	return nil
}
