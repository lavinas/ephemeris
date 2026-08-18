package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/port"
)

// Session CreateRequest represents a request to create a new session.
type SessionCreateRequest struct {
	Nickname string `json:"nickname" validate:"required"`
	Date     string `json:"date" validate:"required"`
	Minutes  int    `json:"minutes" validate:"required"`
	Status   string `json:"status" validate:"required"`
	Comments string `json:"comments,omitempty"`
}

// SessionCreateResponse represents the response after creating a new session.
type SessionCreateResponse struct {
	ResponseBase
	SessionID int64 `json:"session_id"`
}

// NewSessionCreateResponse creates a new instance of SessionCreateResponse with the provided parameters.
func NewSessionCreateResponse(statusCode int, statusMessage string, errorMessage string, sessionID int64) *SessionCreateResponse {
	return &SessionCreateResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		SessionID: sessionID,
	}
}

// Validate checks if the SessionCreateRequest has valid data.
func (r *SessionCreateRequest) Validate(repo port.Repository) error {
	var errs []error
	if err := r.validateNickName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateStatus(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateMinutes(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateNickName checks if the provided nickname is valid.
func (r *SessionCreateRequest) validateNickName() error {
	if r.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	if r.Nickname != strings.ToLower(r.Nickname) {
		return fmt.Errorf("nickname must be lowercase")
	}
	if strings.Contains(r.Nickname, " ") {
		return fmt.Errorf("nickname must not contain spaces")
	}
	// verify if nickname has only letters, numbers, and underscores
	for _, char := range r.Nickname {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("nickname must contain only lowercase letters, numbers, and underscores")
		}
	}
	return nil
}

// validateDate checks if the provided date is in a valid format (YYYY-MM-DD).
func (r *SessionCreateRequest) validateDate() error {
	if r.Date == "" {
		return fmt.Errorf("date is required")
	}
	dtime, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	if dtime.IsZero() {
		return fmt.Errorf("date cannot be zero")
	}

	return nil
}

// validateStatus checks if the provided status is valid.
func (r *SessionCreateRequest) validateStatus() error {
	validStatuses := map[string]bool{
		"realizada":            true,
		"cancelada_cobrar":     true,
		"cancelada_nao_cobrar": true,
	}
	if !validStatuses[r.Status] {
		return fmt.Errorf("invalid status, must be one of: realizada, cancelada_cobrar, cancelada_nao_cobrar")
	}
	return nil
}

// validateMinutes checks if the provided minutes is a positive integer.
func (r *SessionCreateRequest) validateMinutes() error {
	if r.Minutes <= 0 {
		return fmt.Errorf("minutes must be a positive integer")
	}
	return nil
}

// Reset clears the fields of the SessionCreateRequest, setting them to their zero values.
func (r *SessionCreateRequest) Reset() {
	r.Nickname = ""
	r.Date = ""
	r.Minutes = 0
	r.Status = ""
	r.Comments = ""
}
