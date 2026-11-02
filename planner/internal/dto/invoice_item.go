package dto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// InvoiceItemAddRequest represents a request to add a standalone item (addition or discount) to an existing pre-invoice.
type InvoiceItemAddRequest struct {
	InvoiceID   int64   `json:"invoice_id"`
	ServiceID   int64   `json:"service_id"`
	ServiceName string  `json:"service_name,omitempty"`
	Operation   string  `json:"operation,omitempty"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Description string  `json:"description,omitempty"`
}

// UnmarshalJSON custom unmarshaler supporting English and Portuguese/form field names.
func (r *InvoiceItemAddRequest) UnmarshalJSON(data []byte) error {
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

	getFloat64 := func(keys ...string) float64 {
		for _, k := range keys {
			if v, ok := raw[k]; ok && v != nil {
				switch val := v.(type) {
				case float64:
					return val
				case int:
					return float64(val)
				case int64:
					return float64(val)
				case string:
					if f, err := ParsePrice(val); err == nil {
						return f
					}
				}
			}
		}
		return 0
	}

	r.InvoiceID = getInt64("invoice_id", "invoiceId", "fatura_id")
	r.ServiceID = getInt64("service_id", "serviceId", "servico_id")
	r.ServiceName = getString("service_name", "serviceName", "servico", "service")
	r.Operation = getString("operation", "operacao", "tipo_operacao", "type", "tipo")
	r.PeriodStart = getString("period_start", "periodStart", "inicio", "periodo_inicio", "periodo_inicial")
	r.PeriodEnd = getString("period_end", "periodEnd", "fim", "periodo_fim", "periodo_final")
	r.Price = getFloat64("price", "preco", "valor")
	r.Quantity = getInt("quantity", "quantidade", "qtd", "invoiced_quantity")
	r.Description = getString("description", "descricao")

	return nil
}

// ApplyDefaults applies default period dates (current month) and normalizes the price sign based on Operation.
func (r *InvoiceItemAddRequest) ApplyDefaults() {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	if strings.TrimSpace(r.PeriodStart) == "" {
		r.PeriodStart = firstDay.Format("2006-01-02")
	}
	if strings.TrimSpace(r.PeriodEnd) == "" {
		r.PeriodEnd = lastDay.Format("2006-01-02")
	}

	op := strings.ToLower(strings.TrimSpace(r.Operation))
	switch op {
	case "discount", "desconto", "sub", "subtract", "-":
		r.Price = -math.Abs(r.Price)
		r.Operation = "discount"
	case "add", "adicao", "adição", "addition", "+":
		r.Price = math.Abs(r.Price)
		r.Operation = "add"
	default:
		if r.Price < 0 {
			r.Operation = "discount"
		} else if r.Price > 0 {
			r.Operation = "add"
		}
	}
}

// Validate validates the request to add a standalone invoice item.
func (r *InvoiceItemAddRequest) Validate(repo port.Repository) error {
	r.ApplyDefaults()

	if r.InvoiceID <= 0 {
		return errors.New("invoice_id deve ser maior que zero")
	}

	op := strings.ToLower(strings.TrimSpace(r.Operation))
	if op != "" && op != "add" && op != "discount" {
		return fmt.Errorf("operação inválida '%s': esperado 'add' (adição) ou 'discount' (desconto)", r.Operation)
	}

	pStart, err := ParseFlexibleDate(r.PeriodStart)
	if err != nil {
		return fmt.Errorf("período inicial inválido: %w", err)
	}
	pEnd, err := ParseFlexibleDate(r.PeriodEnd)
	if err != nil {
		return fmt.Errorf("período final inválido: %w", err)
	}
	if pEnd.Before(pStart) {
		return errors.New("o período final deve ser maior ou igual ao período inicial")
	}

	if r.Price == 0 {
		return errors.New("o preço deve ser diferente de zero")
	}
	if r.Quantity <= 0 {
		return errors.New("a quantidade deve ser maior que zero")
	}

	if r.ServiceID <= 0 && strings.TrimSpace(r.ServiceName) == "" {
		return errors.New("o serviço deve ser informado (service_id ou service_name)")
	}

	if repo != nil {
		inv, err := repo.GetInvoiceByID(r.InvoiceID)
		if err != nil {
			return fmt.Errorf("erro ao consultar fatura: %w", err)
		}
		if inv == nil || inv.DeletedAt != nil {
			return fmt.Errorf("pré-fatura #%d não encontrada", r.InvoiceID)
		}

		services, err := repo.FindServices(1)
		if err != nil {
			return fmt.Errorf("erro ao buscar serviços: %w", err)
		}
		var foundSvc *domain.Service
		if r.ServiceID > 0 {
			for i := range services {
				if services[i].ID == r.ServiceID {
					foundSvc = &services[i]
					break
				}
			}
		} else if strings.TrimSpace(r.ServiceName) != "" {
			for i := range services {
				if strings.EqualFold(services[i].Name, strings.TrimSpace(r.ServiceName)) {
					foundSvc = &services[i]
					r.ServiceID = services[i].ID
					break
				}
			}
		}
		if foundSvc == nil {
			return errors.New("serviço não encontrado")
		}
	}

	return nil
}

// InvoiceItemAddResponse represents the response after adding a standalone item to an invoice.
type InvoiceItemAddResponse struct {
	ResponseBase
	Invoice *InvoiceDTO     `json:"invoice,omitempty"`
	Item    *InvoiceItemDTO `json:"item,omitempty"`
}

// NewInvoiceItemAddResponse creates a new instance of InvoiceItemAddResponse.
func NewInvoiceItemAddResponse(statusCode int, statusMessage, errorMessage string, invoice *InvoiceDTO, item *InvoiceItemDTO) *InvoiceItemAddResponse {
	return &InvoiceItemAddResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Invoice: invoice,
		Item:    item,
	}
}

// InvoiceItemDeleteRequest represents a request to delete an item from a pre-invoice.
type InvoiceItemDeleteRequest struct {
	ItemID    int64 `json:"id"`
	InvoiceID int64 `json:"invoice_id,omitempty"`
}

// UnmarshalJSON custom unmarshaler supporting English and Portuguese/form field names.
func (r *InvoiceItemDeleteRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
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

	r.ItemID = getInt64("id", "item_id", "itemId", "invoice_item_id", "invoiceItemId")
	r.InvoiceID = getInt64("invoice_id", "invoiceId", "fatura_id")
	return nil
}

// Validate validates the request to delete an invoice item.
func (r *InvoiceItemDeleteRequest) Validate(repo port.Repository) error {
	if r.ItemID <= 0 {
		return errors.New("id do item de fatura deve ser maior que zero")
	}

	if repo != nil {
		if r.InvoiceID > 0 {
			inv, err := repo.GetInvoiceByID(r.InvoiceID)
			if err != nil {
				return fmt.Errorf("erro ao consultar fatura: %w", err)
			}
			if inv == nil || inv.DeletedAt != nil {
				return fmt.Errorf("pré-fatura #%d não encontrada", r.InvoiceID)
			}
			found := false
			for _, it := range inv.Items {
				if it.ID == r.ItemID && it.DeletedAt == nil {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("item #%d não encontrado na pré-fatura #%d", r.ItemID, r.InvoiceID)
			}
		} else if cr, ok := repo.(port.ComptimeRepository); ok {
			it, err := cr.GetInvoiceItemByID(context.Background(), r.ItemID)
			if err != nil {
				return fmt.Errorf("erro ao consultar item de fatura: %w", err)
			}
			if it == nil || it.DeletedAt != nil {
				return fmt.Errorf("item de fatura #%d não encontrado", r.ItemID)
			}
			r.InvoiceID = it.InvoiceID
		}
	}

	return nil
}

// InvoiceItemDeleteResponse represents the response after deleting an item from an invoice.
type InvoiceItemDeleteResponse struct {
	ResponseBase
	Invoice *InvoiceDTO `json:"invoice,omitempty"`
}

// NewInvoiceItemDeleteResponse creates a new instance of InvoiceItemDeleteResponse.
func NewInvoiceItemDeleteResponse(statusCode int, statusMessage, errorMessage string, invoice *InvoiceDTO) *InvoiceItemDeleteResponse {
	return &InvoiceItemDeleteResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Invoice: invoice,
	}
}
