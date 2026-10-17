package dto

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"planner/internal/port"
)

// InvoiceBatchDeleteRequest represents a request to soft-delete multiple invoices matching filter criteria.
type InvoiceBatchDeleteRequest struct {
	StartDate   string  `json:"start_date,omitempty"`
	EndDate     string  `json:"end_date,omitempty"`
	InvoiceID   int64   `json:"invoice_id,omitempty"`
	Nickname    string  `json:"nickname,omitempty"`
	CustomerIDs []int64 `json:"-"`
	PlanType    int     `json:"plan_type,omitempty"`
	ServiceID   int64   `json:"service_id,omitempty"`
	PlanID      int64   `json:"plan_id,omitempty"`
}

// UnmarshalJSON custom unmarshaler supporting both English and Portuguese field names.
func (r *InvoiceBatchDeleteRequest) UnmarshalJSON(data []byte) error {
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

	r.StartDate = getString("start_date", "data_inicio", "inicio")
	r.EndDate = getString("end_date", "data_fim", "fim")
	r.InvoiceID = getInt64("invoice_id", "fatura_id", "id")
	r.PlanID = getInt64("plan_id", "plano_id")
	r.Nickname = getString("nickname", "cliente", "nome")
	r.PlanType = getInt("plan_type", "tipo_plano", "tipo_filtro")
	r.ServiceID = getInt64("service_id", "servico_id", "servico_filtro")

	return nil
}

// InvoiceBatchDeleteResponse represents the response after deleting invoices in batch.
type InvoiceBatchDeleteResponse struct {
	ResponseBase
	DeletedCount int64 `json:"deleted_count"`
}

// NewInvoiceBatchDeleteResponse creates a new instance of InvoiceBatchDeleteResponse.
func NewInvoiceBatchDeleteResponse(statusCode int, statusMessage, errorMessage string, deletedCount int64) *InvoiceBatchDeleteResponse {
	return &InvoiceBatchDeleteResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		DeletedCount: deletedCount,
	}
}

// Validate checks if the InvoiceBatchDeleteRequest has valid filter data.
func (r *InvoiceBatchDeleteRequest) Validate(repo port.Repository) error {
	var errs []error
	if r.InvoiceID < 0 {
		errs = append(errs, errors.New("id da fatura inválido: deve ser positivo"))
	}
	if r.PlanID < 0 {
		errs = append(errs, errors.New("id do plano inválido: deve ser positivo"))
	}
	if r.PlanType < 0 || r.PlanType > 4 {
		errs = append(errs, errors.New("tipo de plano inválido: deve ser entre 0 e 4"))
	}
	if r.ServiceID < 0 {
		errs = append(errs, errors.New("id do serviço inválido: deve ser positivo"))
	}

	var dStart, dEnd time.Time
	if r.StartDate != "" {
		t, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			errs = append(errs, errors.New("formato de data de início inválido: deve ser AAAA-MM-DD"))
		} else {
			dStart = t
		}
	}
	if r.EndDate != "" {
		t, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			errs = append(errs, errors.New("formato de data de fim inválido: deve ser AAAA-MM-DD"))
		} else {
			dEnd = t
		}
	}
	if !dStart.IsZero() && !dEnd.IsZero() && dEnd.Before(dStart) {
		errs = append(errs, errors.New("data de fim não pode ser anterior à data de início"))
	}

	if strings.TrimSpace(r.Nickname) != "" && repo != nil {
		customers, err := repo.FindCustomers(0, 0, 1, nil, &r.Nickname, nil, nil, nil, nil)
		if err != nil {
			errs = append(errs, err)
		} else if len(customers) == 0 {
			r.CustomerIDs = []int64{-1}
		} else {
			r.CustomerIDs = make([]int64, len(customers))
			for i, c := range customers {
				r.CustomerIDs[i] = c.ID
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Reset clears the fields of InvoiceBatchDeleteRequest.
func (r *InvoiceBatchDeleteRequest) Reset() {
	r.StartDate = ""
	r.EndDate = ""
	r.InvoiceID = 0
	r.Nickname = ""
	r.CustomerIDs = nil
	r.PlanType = 0
	r.ServiceID = 0
	r.PlanID = 0
}
