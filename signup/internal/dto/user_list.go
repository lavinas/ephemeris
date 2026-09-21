package dto

import (
	"errors"
	"fmt"
	"strings"

	"signup/internal/port"
	"signup/internal/domain"
)

// UserListRequest represents the request payload for listing users.
type UserListRequest struct {
	Page     int     `json:"page" validate:"required,gt=0"`
	PageSize int     `json:"page_size" validate:"required,gt=0"`
	Vendor   string  `json:"vendor" validate:"required"`
	VendorID int64   `json:"-" validate:"-"`
	Name     *string `json:"name,omitempty"`
	Username *string `json:"username,omitempty"`
	Status   *int    `json:"status,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
}

// UserListResponse represents the response payload for listing users.
type UserListResponse struct {
	Users      []UserListItem `json:"users"`
	TotalCount int            `json:"total_count"`
}

// UserListItem represents a single user in the list response.
type UserListItem struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Whatsapp string  `json:"whatsapp"`
	Status   int     `json:"status"`
}

// NewUserListResponse creates a new UserListResponse with the given users and total count.
func NewUserListResponse(users []UserListItem, totalCount int) *UserListResponse {
	return &UserListResponse{
		Users:      users,
		TotalCount: totalCount,
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
func (r *UserListRequest) Validate(repo port.Repository) error {
	errs := []error{}
	if r.Page <= 0 {
		errs = append(errs, fmt.Errorf("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, fmt.Errorf("page_size must be greater than 0"))
	}
	if r.Vendor == "" {
		errs = append(errs, fmt.Errorf("vendor is required"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
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
func (r *UserListRequest) validateVendor(repo port.Repository) error {
	if r.Vendor == "" {
		return fmt.Errorf("vendor is required")
	}
	vendor := domain.Vendor{}
	if ok, err := vendor.GetByNickname(repo, r.Vendor); err != nil {
		return fmt.Errorf("error fetching vendor by nickname: %w", err)
	} else if !ok {
		return fmt.Errorf("vendor is not valid")
	}
	r.VendorID = vendor.ID
	return nil
}

