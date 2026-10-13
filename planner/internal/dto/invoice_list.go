package dto

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"planner/internal/port"
)

// InvoiceListRequest represents a request to list invoices with filters and pagination.
type InvoiceListRequest struct {
	Page        int     `json:"page"`
	PageSize    int     `json:"page_size"`
	StartDate   string  `json:"start_date,omitempty"`
	EndDate     string  `json:"end_date,omitempty"`
	InvoiceID   int64   `json:"invoice_id,omitempty"`
	Nickname    string  `json:"nickname,omitempty"`
	CustomerIDs []int64 `json:"-"`
	PlanType    int     `json:"plan_type,omitempty"`
	ServiceID   int64   `json:"service_id,omitempty"`
	PlanID      int64   `json:"plan_id,omitempty"`
}

// UnmarshalJSON custom unmarshaler supporting both English and Portuguese/alternative field names.
func (r *InvoiceListRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	getInt := func(keys ...string) int {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					return int(val)
				case int:
					return val
				case int64:
					return int(val)
				case string:
					if i, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
						return i
					}
				}
			}
		}
		return 0
	}

	getInt64 := func(keys ...string) int64 {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					return int64(val)
				case int64:
					return val
				case int:
					return int64(val)
				case string:
					if i, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64); err == nil {
						return i
					}
				}
			}
		}
		return 0
	}

	getString := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				if s, ok := v.(string); ok {
					return strings.TrimSpace(s)
				}
			}
		}
		return ""
	}

	r.Page = getInt("page", "pagina")
	r.PageSize = getInt("page_size", "pageSize", "itens_por_pagina", "itensPorPag")
	r.StartDate = getString("start_date", "startDate", "data_inicio", "date_start", "inicio")
	r.EndDate = getString("end_date", "endDate", "data_fim", "date_end", "fim")
	r.InvoiceID = getInt64("invoice_id", "invoiceId", "id", "fatura_id")
	r.Nickname = getString("nickname", "customer_nickname", "cliente")
	r.PlanType = getInt("plan_type", "planType", "tipo_plano", "tipo_filtro")
	r.ServiceID = getInt64("service_id", "serviceId", "servico_id", "servico_filtro")
	r.PlanID = getInt64("plan_id", "planId", "plano_id")

	return nil
}

// InvoiceListResponse represents the paginated response containing a list of invoices.
type InvoiceListResponse struct {
	ResponseBase
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
	TotalPages  int          `json:"total_pages"`
	TotalCount  int64        `json:"total_count"`
	TotalAmount float64      `json:"total_amount"`
	Invoices    []InvoiceDTO `json:"invoices"`
}

// InvoiceDTO represents an invoice and its summary information.
type InvoiceDTO struct {
	ID          int64            `json:"id"`
	InvoiceDate string           `json:"invoice_date"`
	Nickname    string           `json:"nickname"`
	DueDate     string           `json:"due_date"`
	Amount      float64          `json:"amount"`
	Items       []InvoiceItemDTO `json:"items"`
}

// InvoiceItemDTO represents an item contained within an invoice.
type InvoiceItemDTO struct {
	ID               int64   `json:"id"`
	ServiceID        int64   `json:"service_id"`
	Description      string  `json:"description"`
	Price            float64 `json:"price"`
	InvoicedQuantity int     `json:"invoiced_quantity"`
	RealizedQuantity int     `json:"realized_quantity"`
}

// NewInvoiceListResponse creates a new instance of InvoiceListResponse.
func NewInvoiceListResponse(statusCode int, statusMessage, errorMessage string, invoices []InvoiceDTO,
	page, pageSize, totalPages int, totalCount int64, totalAmount float64) *InvoiceListResponse {
	return &InvoiceListResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		TotalAmount: totalAmount,
		Invoices:    invoices,
	}
}

// Validate checks if the InvoiceListRequest has valid data and applies defaults.
func (r *InvoiceListRequest) Validate(repo port.Repository) error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 10
	}

	var errs []error
	if err := r.validateInvoiceID(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePlanID(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePlanType(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateServiceID(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDates(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *InvoiceListRequest) validateInvoiceID() error {
	if r.InvoiceID < 0 {
		return errors.New("invoice_id não pode ser negativo")
	}
	return nil
}

func (r *InvoiceListRequest) validatePlanID() error {
	if r.PlanID < 0 {
		return errors.New("plan_id não pode ser negativo")
	}
	return nil
}

func (r *InvoiceListRequest) validatePlanType() error {
	if r.PlanType < 0 || r.PlanType > 4 {
		return errors.New("plan_type inválido, deve ser entre 1 e 4")
	}
	return nil
}

func (r *InvoiceListRequest) validateServiceID() error {
	if r.ServiceID < 0 {
		return errors.New("service_id não pode ser negativo")
	}
	return nil
}

func (r *InvoiceListRequest) validateDates() error {
	var sd *time.Time
	if r.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			return errors.New("formato inválido para start_date, esperado AAAA-MM-DD")
		}
		sd = &startDate
	}
	if r.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			return errors.New("formato inválido para end_date, esperado AAAA-MM-DD")
		}
		if sd != nil && endDate.Before(*sd) {
			return errors.New("end_date não pode ser anterior à start_date")
		}
	}
	return nil
}

func (r *InvoiceListRequest) validateNickname(repo port.Repository) error {
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

// Reset clears the fields of InvoiceListRequest.
func (r *InvoiceListRequest) Reset() {
	r.Page = 1
	r.PageSize = 10
	r.StartDate = ""
	r.EndDate = ""
	r.InvoiceID = 0
	r.Nickname = ""
	r.CustomerIDs = nil
	r.PlanType = 0
	r.ServiceID = 0
	r.PlanID = 0
}
