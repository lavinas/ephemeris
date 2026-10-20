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

// InvoiceItemUpdateRequest represents a request to update an item's invoiced quantity.
type InvoiceItemUpdateRequest struct {
	ID               int64 `json:"id"`
	InvoicedQuantity int   `json:"invoiced_quantity"`
}

// InvoiceUpdateRequest represents a request to update an invoice's due date and items' invoiced quantities.
type InvoiceUpdateRequest struct {
	InvoiceID int64                      `json:"id"`
	DueDate   string                     `json:"due_date"`
	Items     []InvoiceItemUpdateRequest `json:"items"`
}

// UnmarshalJSON custom unmarshaler supporting both English and Portuguese/form field names.
func (r *InvoiceUpdateRequest) UnmarshalJSON(data []byte) error {
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

	r.InvoiceID = getInt64("id", "invoice_id", "fatura_id")
	r.DueDate = getString("due_date", "vencimento", "data_vencimento")

	if itemsRaw, ok := raw["items"].([]interface{}); ok {
		r.Items = make([]InvoiceItemUpdateRequest, 0, len(itemsRaw))
		for _, itemVal := range itemsRaw {
			if itemMap, ok := itemVal.(map[string]interface{}); ok {
				var item InvoiceItemUpdateRequest
				if idVal, ok := itemMap["id"]; ok && idVal != nil {
					switch v := idVal.(type) {
					case float64:
						item.ID = int64(v)
					case int64:
						item.ID = v
					case int:
						item.ID = int64(v)
					case string:
						item.ID, _ = strconv.ParseInt(v, 10, 64)
					}
				}
				if qtyVal, ok := itemMap["invoiced_quantity"]; ok && qtyVal != nil {
					switch v := qtyVal.(type) {
					case float64:
						item.InvoicedQuantity = int(v)
					case int:
						item.InvoicedQuantity = v
					case int64:
						item.InvoicedQuantity = int(v)
					case string:
						item.InvoicedQuantity, _ = strconv.Atoi(v)
					}
				}
				r.Items = append(r.Items, item)
			}
		}
	}

	return nil
}

// InvoiceUpdateResponse represents the response after updating an invoice.
type InvoiceUpdateResponse struct {
	ResponseBase
	Invoice *InvoiceDTO `json:"invoice,omitempty"`
}

// NewInvoiceUpdateResponse creates a new instance of InvoiceUpdateResponse.
func NewInvoiceUpdateResponse(statusCode int, statusMessage string, errorMessage string, invoice *InvoiceDTO) *InvoiceUpdateResponse {
	return &InvoiceUpdateResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Invoice: invoice,
	}
}

// Validate checks if the InvoiceUpdateRequest has valid data and ensures only allowed fields are updated.
func (r *InvoiceUpdateRequest) Validate(repo port.Repository) error {
	if r.InvoiceID <= 0 {
		return errors.New("id da pré-fatura inválido: deve ser maior que zero")
	}

	if strings.TrimSpace(r.DueDate) == "" {
		return errors.New("data de vencimento é obrigatória")
	}

	if _, err := time.Parse("2006-01-02", r.DueDate); err != nil {
		return errors.New("formato de data de vencimento inválido: deve ser AAAA-MM-DD")
	}

	if repo != nil {
		inv, err := repo.GetInvoiceByID(r.InvoiceID)
		if err != nil {
			return fmt.Errorf("erro ao buscar pré-fatura: %w", err)
		}
		if inv == nil || inv.DeletedAt != nil {
			return errors.New("pré-fatura não encontrada")
		}

		itemMap := make(map[int64]bool)
		for _, it := range inv.Items {
			if it.DeletedAt == nil {
				itemMap[it.ID] = true
			}
		}

		for _, itemReq := range r.Items {
			if itemReq.ID <= 0 {
				return errors.New("id do item da pré-fatura inválido")
			}
			if !itemMap[itemReq.ID] {
				return fmt.Errorf("item #%d não pertence a esta pré-fatura", itemReq.ID)
			}
			if itemReq.InvoicedQuantity < 0 {
				return fmt.Errorf("quantidade faturada do item #%d não pode ser negativa", itemReq.ID)
			}
		}
	}

	return nil
}
