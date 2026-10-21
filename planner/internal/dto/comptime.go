package dto

import (
	"fmt"
	"time"

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
	VendorID  int64  `json:"vendor_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	DataType  string `json:"data_type"` // "session", "invoice", "all"
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
	if r.DataType == "" {
		r.DataType = "all"
	}
	if r.DataType != "session" && r.DataType != "invoice" && r.DataType != "all" {
		return fmt.Errorf("data_type deve ser 'session', 'invoice' ou 'all'")
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
