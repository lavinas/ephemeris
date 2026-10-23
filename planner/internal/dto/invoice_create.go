package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// InvoiceItemCreateRequest represents an item within a new manual invoice.
type InvoiceItemCreateRequest struct {
	ServiceID   int64   `json:"service_id"`
	ServiceName string  `json:"service_name,omitempty"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
}

// InvoiceCreateRequest represents a request to manually create a new invoice.
type InvoiceCreateRequest struct {
	Nickname           string                     `json:"nickname"`
	InvoiceDate        string                     `json:"invoice_date"`
	DueDate            string                     `json:"due_date"`
	PeriodStart        string                     `json:"period_start"`
	PeriodEnd          string                     `json:"period_end"`
	ReferenceInvoiceID *int64                     `json:"reference_invoice_id,omitempty"`
	Items              []InvoiceItemCreateRequest `json:"items"`
}

// UnmarshalJSON custom unmarshaler supporting English and Portuguese/form field names.
func (r *InvoiceCreateRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
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

	getInt64Ptr := func(keys ...string) *int64 {
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
				case string:
					s := strings.TrimSpace(val)
					if s != "" {
						if i, err := strconv.ParseInt(s, 10, 64); err == nil {
							return &i
						}
					}
				}
			}
		}
		return nil
	}

	r.Nickname = getString("nickname", "cliente", "customer", "customer_nickname")
	r.InvoiceDate = getString("invoice_date", "data_invoice", "faturamento", "data_faturamento", "data")
	r.DueDate = getString("due_date", "vencimento", "data_vencimento")
	r.PeriodStart = getString("period_start", "inicio", "periodo_inicio", "periodo_inicial")
	r.PeriodEnd = getString("period_end", "fim", "periodo_fim", "periodo_final")
	r.ReferenceInvoiceID = getInt64Ptr("reference_invoice_id", "referencia", "reference", "ref_id")

	var itemsRaw []interface{}
	for _, key := range []string{"items", "itens", "invoice_items"} {
		if rawSlice, ok := raw[key].([]interface{}); ok {
			itemsRaw = rawSlice
			break
		}
	}

	if itemsRaw != nil {
		r.Items = make([]InvoiceItemCreateRequest, 0, len(itemsRaw))
		for _, itemVal := range itemsRaw {
			if itemMap, ok := itemVal.(map[string]interface{}); ok {
				var item InvoiceItemCreateRequest
				if s, ok := itemMap["service_name"].(string); ok {
					item.ServiceName = strings.TrimSpace(s)
				} else if s, ok := itemMap["servico"].(string); ok {
					item.ServiceName = strings.TrimSpace(s)
				} else if s, ok := itemMap["service"].(string); ok {
					item.ServiceName = strings.TrimSpace(s)
				}

				if sid, ok := itemMap["service_id"]; ok && sid != nil {
					switch v := sid.(type) {
					case float64:
						item.ServiceID = int64(v)
					case int64:
						item.ServiceID = v
					case int:
						item.ServiceID = int64(v)
					case string:
						item.ServiceID, _ = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					}
				} else if sid, ok := itemMap["servico_id"]; ok && sid != nil {
					switch v := sid.(type) {
					case float64:
						item.ServiceID = int64(v)
					case int64:
						item.ServiceID = v
					case int:
						item.ServiceID = int64(v)
					case string:
						item.ServiceID, _ = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					}
				}

				if p, ok := itemMap["price"]; ok && p != nil {
					switch v := p.(type) {
					case float64:
						item.Price = v
					case int:
						item.Price = float64(v)
					case string:
						clean := strings.ReplaceAll(strings.TrimSpace(v), ",", ".")
						item.Price, _ = strconv.ParseFloat(clean, 64)
					}
				} else if p, ok := itemMap["preco"]; ok && p != nil {
					switch v := p.(type) {
					case float64:
						item.Price = v
					case int:
						item.Price = float64(v)
					case string:
						clean := strings.ReplaceAll(strings.TrimSpace(v), ",", ".")
						item.Price, _ = strconv.ParseFloat(clean, 64)
					}
				}

				if q, ok := itemMap["quantity"]; ok && q != nil {
					switch v := q.(type) {
					case float64:
						item.Quantity = int(v)
					case int:
						item.Quantity = v
					case string:
						item.Quantity, _ = strconv.Atoi(strings.TrimSpace(v))
					}
				} else if q, ok := itemMap["quantidade"]; ok && q != nil {
					switch v := q.(type) {
					case float64:
						item.Quantity = int(v)
					case int:
						item.Quantity = v
					case string:
						item.Quantity, _ = strconv.Atoi(strings.TrimSpace(v))
					}
				}

				if d, ok := itemMap["description"].(string); ok {
					item.Description = strings.TrimSpace(d)
				} else if d, ok := itemMap["descricao"].(string); ok {
					item.Description = strings.TrimSpace(d)
				}

				r.Items = append(r.Items, item)
			}
		}
	}

	return nil
}

// ParseFlexibleDate parses dates formatted as AAAA-MM-DD or DD/MM/AAAA.
func ParseFlexibleDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("data vazia")
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("02/01/2006", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("formato de data inválido: %s (esperado AAAA-MM-DD ou DD/MM/AAAA)", s)
}

// ApplyDefaults applies default dates if not specified:
// invoice_date: today
// due_date: today + 3 days
// period_start: 1st of current month
// period_end: last day of current month
func (r *InvoiceCreateRequest) ApplyDefaults() {
	now := time.Now()
	if strings.TrimSpace(r.InvoiceDate) == "" {
		r.InvoiceDate = now.Format("2006-01-02")
	}
	if strings.TrimSpace(r.DueDate) == "" {
		r.DueDate = now.AddDate(0, 0, 3).Format("2006-01-02")
	}
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDay.AddDate(0, 1, -1)
	if strings.TrimSpace(r.PeriodStart) == "" {
		r.PeriodStart = firstDay.Format("2006-01-02")
	}
	if strings.TrimSpace(r.PeriodEnd) == "" {
		r.PeriodEnd = lastDay.Format("2006-01-02")
	}
}

// Validate checks all validation rules according to adjust1.md.
func (r *InvoiceCreateRequest) Validate(repo port.Repository) error {
	r.ApplyDefaults()

	if strings.TrimSpace(r.Nickname) == "" {
		return errors.New("nickname do cliente é obrigatório")
	}

	invoiceDate, err := ParseFlexibleDate(r.InvoiceDate)
	if err != nil {
		return fmt.Errorf("data da fatura inválida: %w", err)
	}

	dueDate, err := ParseFlexibleDate(r.DueDate)
	if err != nil {
		return fmt.Errorf("data de vencimento inválida: %w", err)
	}

	if dueDate.Before(invoiceDate) {
		return errors.New("a data de vencimento deve ser maior ou igual à data da fatura")
	}

	periodStart, err := ParseFlexibleDate(r.PeriodStart)
	if err != nil {
		return fmt.Errorf("período inicial inválido: %w", err)
	}

	periodEnd, err := ParseFlexibleDate(r.PeriodEnd)
	if err != nil {
		return fmt.Errorf("período final inválido: %w", err)
	}

	if periodEnd.Before(periodStart) {
		return errors.New("o período final deve ser maior ou igual ao período inicial")
	}

	if len(r.Items) == 0 {
		return errors.New("deve-se ter 1 ou mais itens")
	}

	var services []domain.Service
	if repo != nil {
		services, err = repo.FindServices(1)
		if err != nil {
			return fmt.Errorf("erro ao buscar serviços: %w", err)
		}
	}

	for i, item := range r.Items {
		if item.Price <= 0 {
			return fmt.Errorf("item #%d: preço e quantidade devem ser maior que zero", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item #%d: preço e quantidade devem ser maior que zero", i+1)
		}

		if repo != nil {
			var foundSvc *domain.Service
			if item.ServiceID > 0 {
				for _, s := range services {
					if s.ID == item.ServiceID {
						foundSvc = &s
						break
					}
				}
			} else if strings.TrimSpace(item.ServiceName) != "" {
				for _, s := range services {
					if strings.EqualFold(s.Name, strings.TrimSpace(item.ServiceName)) {
						foundSvc = &s
						break
					}
				}
			}

			if foundSvc == nil {
				return fmt.Errorf("item #%d: cada item deve ter um serviço registrado", i+1)
			}
		}
	}

	if repo != nil {
		customer, err := repo.GetCustomer(1, r.Nickname)
		if err != nil {
			return fmt.Errorf("erro ao buscar cliente: %w", err)
		}
		if customer == nil {
			return fmt.Errorf("cliente '%s' não encontrado", r.Nickname)
		}

		exists, err := repo.CheckInvoiceExists(customer.ID, invoiceDate, periodStart, periodEnd, 0)
		if err != nil {
			return fmt.Errorf("erro ao verificar duplicidade de fatura: %w", err)
		}
		if exists {
			return errors.New("não se pode ter mais do que 1 fatura para o mesmo cliente, data de faturamento e período")
		}
	}

	return nil
}

// InvoiceCreateResponse represents the response after manual invoice creation.
type InvoiceCreateResponse struct {
	ResponseBase
	Invoice *InvoiceDTO `json:"invoice,omitempty"`
}

// NewInvoiceCreateResponse creates a new instance of InvoiceCreateResponse.
func NewInvoiceCreateResponse(statusCode int, statusMessage, errorMessage string, invoice *InvoiceDTO) *InvoiceCreateResponse {
	return &InvoiceCreateResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Invoice: invoice,
	}
}
