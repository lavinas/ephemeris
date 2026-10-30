package dto

import (
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// ComptimeListRequest represents the criteria to query and filter comptime records.
type ComptimeListRequest struct {
	VendorID   int64  `json:"vendor_id"`
	CustomerID *int64 `json:"customer_id,omitempty"`
	ServiceID  *int64 `json:"service_id,omitempty"`
	StartMonth string `json:"start_month,omitempty"`
	EndMonth   string `json:"end_month,omitempty"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}

// Validate validates the input parameters for listing comptime.
func (r *ComptimeListRequest) Validate(repo port.Repository) error {
	if r.VendorID <= 0 {
		r.VendorID = 1 // default vendor
	}
	if r.StartMonth != "" {
		if _, err := time.Parse("2006-01", r.StartMonth); err != nil {
			if _, err2 := time.Parse("2006-01-02", r.StartMonth); err2 != nil {
				return fmt.Errorf("start_month inválido, utilize formato YYYY-MM: %w", err)
			}
		}
	}
	if r.EndMonth != "" {
		if _, err := time.Parse("2006-01", r.EndMonth); err != nil {
			if _, err2 := time.Parse("2006-01-02", r.EndMonth); err2 != nil {
				return fmt.Errorf("end_month inválido, utilize formato YYYY-MM: %w", err)
			}
		}
	}
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 50
	}
	return nil
}

// ToFilter converts the request to a port.ComptimeFilter.
func (r *ComptimeListRequest) ToFilter() port.ComptimeFilter {
	filter := port.ComptimeFilter{
		VendorID:   r.VendorID,
		CustomerID: r.CustomerID,
		ServiceID:  r.ServiceID,
		Page:       r.Page,
		PageSize:   r.PageSize,
	}
	if r.StartMonth != "" {
		if t, err := time.Parse("2006-01", r.StartMonth); err == nil {
			filter.StartMonth = &t
		} else if t2, err2 := time.Parse("2006-01-02", r.StartMonth); err2 == nil {
			filter.StartMonth = &t2
		}
	}
	if r.EndMonth != "" {
		if t, err := time.Parse("2006-01", r.EndMonth); err == nil {
			filter.EndMonth = &t
		} else if t2, err2 := time.Parse("2006-01-02", r.EndMonth); err2 == nil {
			filter.EndMonth = &t2
		}
	}
	return filter
}

// ComptimeListResponse represents the response containing comptime records.
type ComptimeListResponse struct {
	ResponseBase
	Items []port.ComptimeView `json:"items"`
	Total int64               `json:"total"`
}

// NewComptimeListResponse creates a new ComptimeListResponse instance.
func NewComptimeListResponse(code int, status, message string, items []port.ComptimeView, total int64) *ComptimeListResponse {
	return &ComptimeListResponse{
		ResponseBase: NewResponseBase(code, status, message),
		Items:        items,
		Total:        total,
	}
}

// ComptimeResyncRequest represents the parameters for triggering a resynchronization job.
type ComptimeResyncRequest struct {
	VendorID   int64  `json:"vendor_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	DataType   string `json:"data_type"` // "session", "invoice", "all"
	CustomerID *int64 `json:"customer_id,omitempty"`
	ServiceID  *int64 `json:"service_id,omitempty"`
}

// Validate validates the resynchronization request parameters.
func (r *ComptimeResyncRequest) Validate(repo port.Repository) error {
	if r.VendorID <= 0 {
		r.VendorID = 1 // default vendor
	}
	if r.StartDate == "" {
		return fmt.Errorf("start_date é obrigatória (formato YYYY-MM-DD)")
	}
	if _, err := time.Parse("2006-01-02", r.StartDate); err != nil {
		return fmt.Errorf("start_date inválida: %w", err)
	}
	if r.EndDate == "" {
		return fmt.Errorf("end_date é obrigatória (formato YYYY-MM-DD)")
	}
	if _, err := time.Parse("2006-01-02", r.EndDate); err != nil {
		return fmt.Errorf("end_date inválida: %w", err)
	}
	if r.DataType == "" || r.DataType == "todos" {
		r.DataType = "all"
	}
	if r.DataType != "session" && r.DataType != "invoice" && r.DataType != "all" {
		return fmt.Errorf("data_type deve ser 'session', 'invoice' ou 'all'")
	}
	if r.CustomerID != nil && *r.CustomerID <= 0 {
		r.CustomerID = nil
	}
	if r.ServiceID != nil && *r.ServiceID <= 0 {
		r.ServiceID = nil
	}
	return nil
}

// ComptimeResyncResponse represents the output of a resynchronization job.
type ComptimeResyncResponse struct {
	ResponseBase
	Result *port.ResyncResult `json:"result,omitempty"`
}

// NewComptimeResyncResponse creates a new ComptimeResyncResponse instance.
func NewComptimeResyncResponse(code int, status, message string, result *port.ResyncResult) *ComptimeResyncResponse {
	return &ComptimeResyncResponse{
		ResponseBase: NewResponseBase(code, status, message),
		Result:       result,
	}
}

// ComptimeDeleteRequest represents the parameters for deleting a comptime record.
type ComptimeDeleteRequest struct {
	ID int64 `json:"id"`
}

// Validate validates the delete request.
func (r *ComptimeDeleteRequest) Validate() error {
	if r.ID <= 0 {
		return fmt.Errorf("id é obrigatório e deve ser maior que zero")
	}
	return nil
}

// ComptimeAdjustRequest represents the parameters for creating or updating a manual comptime adjustment.
type ComptimeAdjustRequest struct {
	ID              int64  `json:"id,omitempty"`
	ComptimeID      int64  `json:"comptime_id,omitempty"`
	VendorID        int64  `json:"vendor_id,omitempty"`
	CustomerID      int64  `json:"customer_id,omitempty"`
	Nickname        string `json:"nickname,omitempty"`
	ServiceID       int64  `json:"service_id,omitempty"`
	Month           string `json:"month,omitempty"`
	AdjustedMinutes int    `json:"adjusted_minutes"`
	Description     string `json:"description"`
}

// Validate validates the manual adjustment request.
func (r *ComptimeAdjustRequest) Validate(repo port.Repository) error {
	if r.VendorID <= 0 {
		r.VendorID = 1
	}
	if r.ComptimeID <= 0 && r.ID > 0 {
		r.ComptimeID = r.ID
	}
	if r.ComptimeID <= 0 {
		if r.CustomerID <= 0 && strings.TrimSpace(r.Nickname) != "" && repo != nil {
			if cust, err := repo.GetCustomer(r.VendorID, strings.TrimSpace(r.Nickname)); err == nil && cust != nil {
				r.CustomerID = cust.ID
			}
		}
		if r.CustomerID <= 0 {
			return fmt.Errorf("customer_id é obrigatório")
		}
		if r.ServiceID <= 0 {
			return fmt.Errorf("service_id é obrigatório")
		}
		if strings.TrimSpace(r.Month) == "" {
			return fmt.Errorf("month é obrigatório (formato YYYY-MM ou YYYY-MM-DD)")
		}
		if _, err := r.ParsedMonth(); err != nil {
			return err
		}
	}
	r.Description = strings.TrimSpace(r.Description)
	if r.AdjustedMinutes != 0 && r.Description == "" {
		return fmt.Errorf("description (justificativa) é obrigatória")
	}
	if len(r.Description) > 250 {
		return fmt.Errorf("description deve ter no máximo 250 caracteres")
	}
	return nil
}

// ParsedMonth parses the Month string into a normalized time.Time.
func (r *ComptimeAdjustRequest) ParsedMonth() (time.Time, error) {
	m := strings.TrimSpace(r.Month)
	if t, err := time.Parse("2006-01", m); err == nil {
		return domain.NormalizeMonth(t), nil
	}
	if t, err := time.Parse("2006-01-02", m); err == nil {
		return domain.NormalizeMonth(t), nil
	}
	return time.Time{}, fmt.Errorf("month inválido, utilize formato YYYY-MM ou YYYY-MM-DD")
}

// ComptimeAdjustDeleteRequest represents the parameters for deleting a manual comptime adjustment.
type ComptimeAdjustDeleteRequest struct {
	ID         int64  `json:"id,omitempty"`
	ComptimeID int64  `json:"comptime_id,omitempty"`
	VendorID   int64  `json:"vendor_id,omitempty"`
	CustomerID int64  `json:"customer_id,omitempty"`
	Nickname   string `json:"nickname,omitempty"`
	ServiceID  int64  `json:"service_id,omitempty"`
	Month      string `json:"month,omitempty"`
}

// Validate validates the adjustment delete request.
func (r *ComptimeAdjustDeleteRequest) Validate(repo port.Repository) error {
	if r.VendorID <= 0 {
		r.VendorID = 1
	}
	if r.ComptimeID <= 0 && r.ID > 0 {
		r.ComptimeID = r.ID
	}
	if r.ComptimeID <= 0 {
		if r.CustomerID <= 0 && strings.TrimSpace(r.Nickname) != "" && repo != nil {
			if cust, err := repo.GetCustomer(r.VendorID, strings.TrimSpace(r.Nickname)); err == nil && cust != nil {
				r.CustomerID = cust.ID
			}
		}
		if r.CustomerID <= 0 {
			return fmt.Errorf("id ou (customer_id, service_id, month) são obrigatórios")
		}
		if r.ServiceID <= 0 {
			return fmt.Errorf("service_id é obrigatório")
		}
		if strings.TrimSpace(r.Month) == "" {
			return fmt.Errorf("month é obrigatório (formato YYYY-MM ou YYYY-MM-DD)")
		}
		if _, err := r.ParsedMonth(); err != nil {
			return err
		}
	}
	return nil
}

// ParsedMonth parses the Month string into a normalized time.Time.
func (r *ComptimeAdjustDeleteRequest) ParsedMonth() (time.Time, error) {
	m := strings.TrimSpace(r.Month)
	if t, err := time.Parse("2006-01", m); err == nil {
		return domain.NormalizeMonth(t), nil
	}
	if t, err := time.Parse("2006-01-02", m); err == nil {
		return domain.NormalizeMonth(t), nil
	}
	return time.Time{}, fmt.Errorf("month inválido, utilize formato YYYY-MM ou YYYY-MM-DD")
}

// ComptimeAdjustResponse represents the response after saving or deleting a comptime adjustment.
type ComptimeAdjustResponse struct {
	ResponseBase
	Comptime *domain.Comptime       `json:"comptime,omitempty"`
	Adjust   *domain.ComptimeAdjust `json:"adjust,omitempty"`
}

// NewComptimeAdjustResponse creates a new ComptimeAdjustResponse instance.
func NewComptimeAdjustResponse(code int, status, message string, comptime *domain.Comptime, adjust *domain.ComptimeAdjust) *ComptimeAdjustResponse {
	return &ComptimeAdjustResponse{
		ResponseBase: NewResponseBase(code, status, message),
		Comptime:     comptime,
		Adjust:       adjust,
	}
}
