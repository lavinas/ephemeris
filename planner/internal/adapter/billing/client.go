package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"planner/internal/port"
)

// Client is an HTTP adapter that implements port.BillingClient to communicate with the billing subsystem.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     port.Logger
}

// NewClient creates a new instance of the billing HTTP client adapter.
func NewClient(baseURL string, logger port.Logger, httpClient ...*http.Client) *Client {
	cleanURL := strings.TrimRight(baseURL, "/")
	if cleanURL == "" {
		cleanURL = "http://localhost:8081"
	}
	var client *http.Client
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	} else {
		client = &http.Client{
			Timeout: 15 * time.Second,
		}
	}
	return &Client{
		baseURL:    cleanURL,
		httpClient: client,
		logger:     logger,
	}
}

type billingItemPayload struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type billingInvoicePayload struct {
	Vendor       string               `json:"vendor"`
	Customer     string               `json:"customer"`
	InvoiceDate  string               `json:"invoicing"`
	DueDate      string               `json:"due"`
	ExternalCode *string              `json:"external_code,omitempty"`
	Items        []billingItemPayload `json:"items"`
}

type billingCreateRequestPayload struct {
	Items []billingInvoicePayload `json:"items"`
}

type billingResponseBase struct {
	HttpCode int     `json:"http_code"`
	Status   string  `json:"status"`
	Message  string  `json:"message"`
	ID       int64   `json:"id,omitempty"`
	IDs      []int64 `json:"ids,omitempty"`
}

type billingListResponse struct {
	HttpCode int    `json:"http_code"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Invoices []struct {
		ID           int64  `json:"id"`
		Customer     string `json:"customer"`
		ExternalCode string `json:"external_code"`
	} `json:"invoices"`
}

// CreateInvoice sends an invoice to billing POST /invoice/api/create and returns the created invoice ID.
// If billing reports a duplicate external_code (synchronization race/retry), it attempts to resolve the existing ID.
func (c *Client) CreateInvoice(ctx context.Context, input port.BillingInvoiceInput) (int64, error) {
	if c.logger != nil {
		c.logger.IPrintf(2, "Sending invoice to billing for customer %s, vendor %s", input.Customer, input.Vendor)
	}

	itemsPayload := make([]billingItemPayload, len(input.Items))
	for i, it := range input.Items {
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		itemsPayload[i] = billingItemPayload{
			Description: it.Description,
			Quantity:    qty,
			Price:       it.Price,
		}
	}

	reqBody := billingCreateRequestPayload{
		Items: []billingInvoicePayload{
			{
				Vendor:       input.Vendor,
				Customer:     input.Customer,
				InvoiceDate:  input.Invoicing,
				DueDate:      input.Due,
				ExternalCode: input.ExternalCode,
				Items:        itemsPayload,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("error serializing billing invoice request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/invoice/api/create", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("error creating billing http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("error executing billing request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading billing response: %w", err)
	}

	var respPayload billingResponseBase
	_ = json.Unmarshal(bodyBytes, &respPayload)

	// Check if successful
	if resp.StatusCode == http.StatusOK || (respPayload.HttpCode == http.StatusOK && respPayload.Status == "success") {
		if respPayload.ID > 0 {
			return respPayload.ID, nil
		}
		if len(respPayload.IDs) > 0 && respPayload.IDs[0] > 0 {
			return respPayload.IDs[0], nil
		}
		return 0, errors.New("billing returned success without invoice id")
	}

	// Check for duplicate external code in message or body
	errMsg := respPayload.Message
	if errMsg == "" {
		errMsg = string(bodyBytes)
	}

	isDuplicate := strings.Contains(strings.ToLower(errMsg), "duplicate external_code") ||
		strings.Contains(strings.ToLower(errMsg), "duplicidade")

	if isDuplicate && input.ExternalCode != nil && *input.ExternalCode != "" {
		if c.logger != nil {
			c.logger.IPrintf(1, "Duplicate external_code detected (%s). Fetching existing invoice ID from billing...", *input.ExternalCode)
		}
		existingID, getErr := c.GetInvoiceByExternalCode(ctx, input.Vendor, *input.ExternalCode)
		if getErr == nil && existingID != nil && *existingID > 0 {
			if c.logger != nil {
				c.logger.IPrintf(1, "Resolved existing invoice ID %d for external_code %s", *existingID, *input.ExternalCode)
			}
			return *existingID, nil
		}
	}

	return 0, fmt.Errorf("billing error (%d): %s", resp.StatusCode, errMsg)
}

// GetInvoiceByExternalCode queries billing GET /invoice/api/list to find an invoice by vendor and external code.
func (c *Client) GetInvoiceByExternalCode(ctx context.Context, vendor, externalCode string) (*int64, error) {
	apiURL := fmt.Sprintf("%s/invoice/api/list?vendor=%s&external_code=%s&page=1&page_size=10",
		c.baseURL, url.QueryEscape(vendor), url.QueryEscape(externalCode))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating billing list request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error querying billing invoices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("billing returned status %d: %s", resp.StatusCode, string(body))
	}

	var listResp billingListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("error decoding billing list response: %w", err)
	}

	for _, inv := range listResp.Invoices {
		if inv.ExternalCode == externalCode && inv.ID > 0 {
			return &inv.ID, nil
		}
	}

	return nil, nil
}
