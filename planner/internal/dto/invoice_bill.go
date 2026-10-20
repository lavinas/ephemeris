package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"planner/internal/port"
)

// InvoiceBillRequest represents a request to bill one or multiple pre-invoices.
type InvoiceBillRequest struct {
	InvoiceIDs  []int64 `json:"invoice_ids,omitempty"`
	InvoiceID   int64   `json:"invoice_id,omitempty"`
	Nickname    string  `json:"nickname,omitempty"`
	CustomerIDs []int64 `json:"-"`
	PlanType    int     `json:"plan_type,omitempty"`
	ServiceID   int64   `json:"service_id,omitempty"`
	PlanID      int64   `json:"plan_id,omitempty"`
	StartDate   string  `json:"start_date,omitempty"`
	EndDate     string  `json:"end_date,omitempty"`
}

// UnmarshalJSON custom unmarshaler supporting both English and Portuguese/alternative field names.
func (r *InvoiceBillRequest) UnmarshalJSON(data []byte) error {
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

	getInt64Slice := func(keys ...string) []int64 {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case []interface{}:
					var result []int64
					for _, item := range val {
						switch itemVal := item.(type) {
						case float64:
							result = append(result, int64(itemVal))
						case int64:
							result = append(result, itemVal)
						case int:
							result = append(result, int64(itemVal))
						case string:
							if id, err := strconv.ParseInt(strings.TrimSpace(itemVal), 10, 64); err == nil {
								result = append(result, id)
							}
						}
					}
					return result
				case []int64:
					return val
				case string:
					parts := strings.Split(val, ",")
					var result []int64
					for _, p := range parts {
						if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil && id > 0 {
							result = append(result, id)
						}
					}
					return result
				}
			}
		}
		return nil
	}

	r.InvoiceID = getInt64("invoice_id", "invoiceId", "id", "fatura_id")
	r.InvoiceIDs = getInt64Slice("invoice_ids", "invoiceIds", "ids", "faturas_ids")
	r.StartDate = getString("start_date", "startDate", "data_inicio", "date_start", "inicio")
	r.EndDate = getString("end_date", "endDate", "data_fim", "date_end", "fim")
	r.Nickname = getString("nickname", "customer_nickname", "cliente")
	r.PlanType = getInt("plan_type", "planType", "tipo_plano", "tipo_filtro")
	r.ServiceID = getInt64("service_id", "serviceId", "servico_id", "servico_filtro")
	r.PlanID = getInt64("plan_id", "planId", "plano_id")

	return nil
}

// Validate verifies whether InvoiceBillRequest is valid.
func (r *InvoiceBillRequest) Validate(repo port.Repository) error {
	var errs []error

	if r.InvoiceID < 0 {
		errs = append(errs, errors.New("invoice_id não pode ser negativo"))
	}
	for _, id := range r.InvoiceIDs {
		if id <= 0 {
			errs = append(errs, fmt.Errorf("id de fatura inválido na lista: %d", id))
		}
	}
	if r.PlanID < 0 {
		errs = append(errs, errors.New("plan_id não pode ser negativo"))
	}
	if r.ServiceID < 0 {
		errs = append(errs, errors.New("service_id não pode ser negativo"))
	}

	var dStart, dEnd time.Time
	var errStart, errEnd error
	if r.StartDate != "" {
		dStart, errStart = time.Parse("2006-01-02", r.StartDate)
		if errStart != nil {
			errs = append(errs, errors.New("formato de data_inicio inválido, esperado YYYY-MM-DD"))
		}
	}
	if r.EndDate != "" {
		dEnd, errEnd = time.Parse("2006-01-02", r.EndDate)
		if errEnd != nil {
			errs = append(errs, errors.New("formato de data_fim inválido, esperado YYYY-MM-DD"))
		}
	}
	if errStart == nil && errEnd == nil && !dStart.IsZero() && !dEnd.IsZero() && dEnd.Before(dStart) {
		errs = append(errs, errors.New("data_fim não pode ser anterior a data_inicio"))
	}

	if r.Nickname != "" && repo != nil {
		customers, err := repo.FindCustomers(1, 100, 0, nil, &r.Nickname, nil, nil, nil, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("erro ao buscar cliente por nickname: %w", err))
		} else if len(customers) == 0 {
			r.CustomerIDs = []int64{-1}
		} else {
			for _, c := range customers {
				r.CustomerIDs = append(r.CustomerIDs, c.ID)
			}
		}
	}

	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// InvoiceBillResultItem represents the billing result of an individual invoice.
type InvoiceBillResultItem struct {
	InvoiceID          int64  `json:"invoice_id"`
	ReferenceInvoiceID int64  `json:"reference_invoice_id,omitempty"`
	Status             string `json:"status"`
	Error              string `json:"error,omitempty"`
}

// InvoiceBillResponse represents the result of the invoice billing operation.
type InvoiceBillResponse struct {
	ResponseBase
	BilledCount int                     `json:"billed_count"`
	FailedCount int                     `json:"failed_count"`
	Results     []InvoiceBillResultItem `json:"results,omitempty"`
}

// NewInvoiceBillResponse creates a new instance of InvoiceBillResponse.
func NewInvoiceBillResponse(statusCode int, status, message string, results []InvoiceBillResultItem) *InvoiceBillResponse {
	billed := 0
	failed := 0
	for _, res := range results {
		if res.Status == "success" {
			billed++
		} else {
			failed++
		}
	}
	return &InvoiceBillResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   status,
			Message:  message,
		},
		BilledCount: billed,
		FailedCount: failed,
		Results:     results,
	}
}
