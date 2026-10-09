package dto

import (
	"errors"
	"strings"
	"time"

	"planner/internal/port"
)

// SessionListRequest represents a request to retrieve a list of sessions.
type SessionListRequest struct {
	Page        int     `json:"page" validate:"required,gt=0"`
	PageSize    int     `json:"page_size" validate:"required,gt=0"`
	SessionID   int64   `json:"session_id,omitempty"`
	Nickname    string  `json:"nickname,omitempty"`
	CustomerIDs []int64 `json:"-"`
	DateStart   string  `json:"date_start,omitempty"`
	DateEnd     string  `json:"date_end,omitempty"`
	Minutes     int     `json:"minutes,omitempty"`
	ServiceID   int64   `json:"service_id,omitempty"`
	Status      string  `json:"status,omitempty"`
	Comments    string  `json:"comments,omitempty"`
}

// SessionListResponse represents the response containing a list of sessions.
type SessionListResponse struct {
	ResponseBase
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
	Sessions   []Session `json:"sessions"`
}

// Session represents a session with its details.
type Session struct {
	SessionID int64  `json:"session_id"`
	Nickname  string `json:"nickname"`
	Date      string `json:"date"`
	Minutes   int    `json:"minutes"`
	Status    string `json:"status"`
	ServiceID int64  `json:"service_id,omitempty"`
	Service   string `json:"service,omitempty"`
	Comments  string `json:"comments,omitempty"`
}

// NewSessionListResponse creates a new instance of SessionListResponse with the provided parameters.
func NewSessionListResponse(statusCode int, statusMessage string, errorMessage string, sessions []Session,
	page int, pageSize int, totalPages int) *SessionListResponse {
	return &SessionListResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Sessions:   sessions,
	}
}

// Validate checks if the SessionListRequest has valid data.
func (r *SessionListRequest) Validate(repo port.Repository) error {
	errs := []error{}
	if err := r.validateSessionID(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDates(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePagination(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateMinutes(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateServiceID(); err != nil {
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

// validateNickname searches for customer IDs matching the nickname pattern for vendor 1.
func (r *SessionListRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		r.CustomerIDs = nil
		return nil
	}
	if repo == nil {
		return nil
	}
	customers, err := repo.FindCustomers(0, 0, 1, nil, &r.Nickname, nil, nil, nil, nil)
	if err != nil {
		return err
	}
	if len(customers) == 0 {
		r.CustomerIDs = []int64{-1}
		return nil
	}
	r.CustomerIDs = make([]int64, len(customers))
	for i, c := range customers {
		r.CustomerIDs[i] = c.ID
	}
	return nil
}

// validateSessionID checks if the provided session ID is valid.
func (r *SessionListRequest) validateSessionID() error {
	if r.SessionID < 0 {
		return errors.New("session_id cannot be negative")
	}
	return nil
}

// validatePagination checks if the provided page and page size values are valid.
func (r *SessionListRequest) validatePagination() error {
	errs := []error{}
	if r.Page <= 0 {
		errs = append(errs, errors.New("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, errors.New("page_size must be greater than 0"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateStartDate checks if the provided start date is valid.
func (r *SessionListRequest) validateDates() error {
	errs := []error{}
	var sd *time.Time
	if r.DateStart != "" {
		startDate, err := time.Parse("2006-01-02", r.DateStart)
		if err != nil {
			errs = append(errs, errors.New("invalid start date format, expected YYYY-MM-DD"))
		} else {
			sd = &startDate
		}
	}
	if r.DateEnd != "" {
		endDate, err := time.Parse("2006-01-02", r.DateEnd)
		if err != nil {
			errs = append(errs, errors.New("invalid end date format, expected YYYY-MM-DD"))
		} else if sd != nil && endDate.Before(*sd) {
			errs = append(errs, errors.New("end date cannot be before start date"))
		}
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// validateMinutes checks if the provided minutes value is valid.
func (r *SessionListRequest) validateMinutes() error {
	if r.Minutes < 0 {
		return errors.New("minutes cannot be negative")
	}
	return nil
}

// validateServiceID checks if the provided service ID is valid.
func (r *SessionListRequest) validateServiceID() error {
	if r.ServiceID < 0 {
		return errors.New("service_id cannot be negative")
	}
	return nil
}

// validateStatus checks if the provided status is valid.
func (r *SessionListRequest) validateStatus() error {
	if r.Status != "" && !validStatuses[r.Status] {
		return errors.New("invalid status value")
	}
	return nil
}

// Reset clears the fields of the SessionListRequest, resetting it to its default state.
func (r *SessionListRequest) Reset() {
	r.Nickname = ""
	r.CustomerIDs = nil
	r.DateStart = ""
	r.DateEnd = ""
	r.Minutes = 0
	r.ServiceID = 0
	r.Status = ""
	r.Comments = ""
}
