package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// InvoicingRequest represents the request DTO for generating invoices from plans.
type InvoicingRequest struct {
	VendorID   *int64  `json:"vendor_id,omitempty"`
	StartDate  *string `json:"start_date,omitempty"`
	EndDate    *string `json:"end_date,omitempty"`
	PlanType   *int    `json:"plan_type,omitempty"`
	ServiceID  *int64  `json:"service_id,omitempty"`
	CustomerID *int64  `json:"customer_id,omitempty"`
	Reprocess  *bool   `json:"reprocess,omitempty"`
	NMonths    *int    `json:"n_months,omitempty"`
}

// UnmarshalJSON custom unmarshaler supporting both English and Portuguese field names.
func (r *InvoicingRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	getInt64 := func(keys ...string) *int64 {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					i := int64(val)
					return &i
				case int64:
					return &val
				case int:
					i := int64(val)
					return &i
				}
			}
		}
		return nil
	}

	getInt := func(keys ...string) *int {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					i := int(val)
					return &i
				case int:
					return &val
				}
			}
		}
		return nil
	}

	getString := func(keys ...string) *string {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					trimmed := strings.TrimSpace(s)
					return &trimmed
				}
			}
		}
		return nil
	}

	getBool := func(keys ...string) *bool {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				if b, ok := v.(bool); ok {
					return &b
				}
			}
		}
		return nil
	}

	r.VendorID = getInt64("vendor_id", "vendor")
	r.CustomerID = getInt64("customer_id", "cliente_id", "customer")
	r.ServiceID = getInt64("service_id", "tipo_servico", "service")
	r.PlanType = getInt("plan_type", "tipo_plano", "type")
	r.NMonths = getInt("n_months", "n_meses")
	r.StartDate = getString("start_date", "data_inicio", "startDate", "inicio")
	r.EndDate = getString("end_date", "data_fim", "endDate", "fim")
	r.Reprocess = getBool("reprocess", "reprocessar")

	return nil
}

// Validate checks if the InvoicingRequest has valid data.
func (r *InvoicingRequest) Validate(repo port.Repository) error {
	if r.PlanType != nil && *r.PlanType != 0 {
		if *r.PlanType < int(domain.PlanTypeAgenda) || *r.PlanType > int(domain.PlanTypePayPerView) {
			return fmt.Errorf("tipo de plano inválido: %d (deve ser entre 1 e 4)", *r.PlanType)
		}
	}

	if r.StartDate != nil && *r.StartDate != "" {
		if _, err := time.Parse("2006-01-02", *r.StartDate); err != nil {
			return errors.New("data de início inválida: formato esperado YYYY-MM-DD")
		}
	}

	if r.EndDate != nil && *r.EndDate != "" {
		if _, err := time.Parse("2006-01-02", *r.EndDate); err != nil {
			return errors.New("data de fim inválida: formato esperado YYYY-MM-DD")
		}
	}

	if r.StartDate != nil && *r.StartDate != "" && r.EndDate != nil && *r.EndDate != "" {
		dStart, _ := time.Parse("2006-01-02", *r.StartDate)
		dEnd, _ := time.Parse("2006-01-02", *r.EndDate)
		if dStart.After(dEnd) {
			return errors.New("data de início não pode ser posterior à data de fim")
		}
	}

	if r.CustomerID != nil && *r.CustomerID > 0 && repo != nil {
		c, err := repo.GetCustomerByID(*r.CustomerID)
		if err != nil {
			return fmt.Errorf("erro ao verificar cliente: %w", err)
		}
		if c == nil {
			return fmt.Errorf("cliente com id %d não encontrado", *r.CustomerID)
		}
	}

	if r.ServiceID != nil && *r.ServiceID > 0 && repo != nil {
		s, err := repo.GetServiceByID(*r.ServiceID)
		if err != nil {
			return fmt.Errorf("erro ao verificar serviço: %w", err)
		}
		if s == nil {
			return fmt.Errorf("serviço com id %d não encontrado", *r.ServiceID)
		}
	}

	return nil
}

// Reset resets all fields in InvoicingRequest.
func (r *InvoicingRequest) Reset() {
	r.VendorID = nil
	r.StartDate = nil
	r.EndDate = nil
	r.PlanType = nil
	r.ServiceID = nil
	r.CustomerID = nil
	r.Reprocess = nil
	r.NMonths = nil
}

// GetParsedDates resolves start and end dates with business defaults:
// - StartDate default: N months before the start of the current month (default N = 1)
// - EndDate default: end of current month
func (r *InvoicingRequest) GetParsedDates(now time.Time) (time.Time, time.Time, error) {
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	currentMonthEnd := currentMonthStart.AddDate(0, 1, -1)

	var startDate time.Time
	if r.StartDate != nil && *r.StartDate != "" {
		d, err := time.Parse("2006-01-02", *r.StartDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		startDate = d
	} else {
		nMonths := 1
		if r.NMonths != nil && *r.NMonths >= 0 {
			nMonths = *r.NMonths
		}
		startDate = currentMonthStart.AddDate(0, -nMonths, 0)
	}

	var endDate time.Time
	if r.EndDate != nil && *r.EndDate != "" {
		d, err := time.Parse("2006-01-02", *r.EndDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		endDate = d
	} else {
		endDate = currentMonthEnd
	}

	return startDate, endDate, nil
}

// IsReprocess returns true if reprocess option is enabled.
func (r *InvoicingRequest) IsReprocess() bool {
	if r.Reprocess != nil {
		return *r.Reprocess
	}
	return false
}

// InvoicingResponse represents the standard base response for invoicing.
type InvoicingResponse struct {
	ResponseBase
}

// NewInvoicingResponse creates a new instance of InvoicingResponse.
func NewInvoicingResponse(httpCode int, status, message string) *InvoicingResponse {
	return &InvoicingResponse{
		ResponseBase: ResponseBase{
			HttpCode: httpCode,
			Status:   status,
			Message:  message,
		},
	}
}
