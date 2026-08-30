package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// SessionUpdateRequest represents a request to update an existing session.
type SessionUpdateRequest struct {
	ID       int64          `json:"id" validate:"required"`
	Session  domain.Session `json:"-"`
	Nickname string         `json:"nickname" validate:"required"`
	Date     string         `json:"date" validate:"required"`
	Minutes  int            `json:"minutes" validate:"required"`
	Service  string         `json:"service" validate:"required"`
	Status   string         `json:"status" validate:"required"`
	Comments string         `json:"comments,omitempty"`
}

// SessionUpdateResponse represents the response after updating an existing session.
type SessionUpdateResponse struct {
	ResponseBase
}

// NewSessionUpdateResponse creates a new instance of SessionUpdateResponse with the provided parameters.
func NewSessionUpdateResponse(statusCode int, statusMessage string, errorMessage string) *SessionUpdateResponse {
	return &SessionUpdateResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
	}
}

// Validate checks if the SessionUpdateRequest has valid data.
func (r *SessionUpdateRequest) Validate(repo port.Repository) error {
	errs := []error{}
	if err := r.validateID(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDate(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateService(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateStatus(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateMinutes(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAlmostOneField(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %v", errs)
	}
	return nil
}

// GetSession returns the Session associated with the SessionUpdateRequest.
func (r *SessionUpdateRequest) GetSession() domain.Session {
	return r.Session
}

// validateID checks if the ID field of the SessionUpdateRequest is valid.
func (r *SessionUpdateRequest) validateID(repo port.Repository) error {
	if r.ID <= 0 {
		return errors.New("invalid session_id: must be greater than 0")
	}
	domainSession := domain.Session{}
	sessions, _, err := domainSession.Find(repo, 1, 1, r.ID, "", time.Time{}, time.Time{}, 0, "", "", "")
	if err != nil {
		return fmt.Errorf("error finding session: %v", err)
	}
	if len(sessions) == 0 {
		return fmt.Errorf("session with id %d not found", r.ID)
	}
	r.Session = sessions[0]
	return nil
}

// validateNickName checks if the provided nickname is valid.
func (r *SessionUpdateRequest) validateNickName() error {
	if r.Nickname == "" {
		return nil
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
func (r *SessionUpdateRequest) validateDate() error {
	if r.Date == "" {
		return nil
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
func (r *SessionUpdateRequest) validateStatus() error {
	if r.Status == "" {
		return nil
	}
	if !validStatuses[r.Status] {
		return fmt.Errorf("invalid status, must be one of: realizada, cancelada_cobrar, cancelada_nao_cobrar")
	}
	return nil
}

// validateService checks if the provided service is valid.
func (r *SessionUpdateRequest) validateService() error {
	if r.Service == "" {
		return nil
	}
	if !validServices[r.Service] {
		return fmt.Errorf("invalid service, must be one of: aula/canto, aula/piano")
	}
	return nil
}

// validateMinutes checks if the provided minutes is a positive integer.
func (r *SessionUpdateRequest) validateMinutes() error {
	if r.Minutes == 0 {
		return nil
	}
	if r.Minutes < 0 {
		return fmt.Errorf("minutes must be a positive integer")
	}
	return nil
}

// validatAlmostOneField checks if at least one of the fields (Nickname, Date, Minutes, Service, Status, Comments) is provided for update.
func (r *SessionUpdateRequest) validateAlmostOneField() error {
	if r.Nickname == "" && r.Date == "" && r.Minutes == 0 && r.Service == "" && r.Status == "" && r.Comments == "" {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// Reset resets the SessionUpdateRequest fields to their zero values.
func (r *SessionUpdateRequest) Reset() {
	r.ID = 0
	r.Nickname = ""
	r.Date = ""
	r.Minutes = 0
	r.Service = ""
	r.Status = ""
	r.Comments = ""
}
