package pluggy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu        sync.RWMutex
	apiKey    string
	expiresAt time.Time
}

func NewClient(baseURL, clientID, clientSecret string) *Client {
	if baseURL == "" {
		baseURL = "https://api.pluggy.ai"
	}
	return &Client{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getAPIKey garante um token válido, renovando se necessário
func (c *Client) getAPIKey(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.apiKey != "" && time.Now().Before(c.expiresAt.Add(-5*time.Minute)) {
		token := c.apiKey
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check
	if c.apiKey != "" && time.Now().Before(c.expiresAt.Add(-5*time.Minute)) {
		return c.apiKey, nil
	}

	authBody, err := json.Marshal(map[string]string{
		"clientId":     c.clientID,
		"clientSecret": c.clientSecret,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao serializar credenciais auth: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth", bytes.NewReader(authBody))
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição de auth: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha na requisição de autenticação Pluggy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("falha de autenticação Pluggy HTTP %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta de auth: %w", err)
	}

	c.apiKey = authResp.APIKey
	// Pluggy API keys normalmente duram 2 horas
	c.expiresAt = time.Now().Add(110 * time.Minute)
	return c.apiKey, nil
}

// doRequest executa uma requisição HTTP autenticada com X-API-KEY
func (c *Client) doRequest(ctx context.Context, method, path string, body any, out any) error {
	apiKey, err := c.getAPIKey(ctx)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("erro ao serializar corpo da requisição: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição para %s: %w", reqURL, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao executar requisição %s %s: %w", method, reqURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro na API Pluggy [%s %s] status %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("erro ao decodificar resposta JSON de %s: %w", path, err)
		}
	}

	return nil
}

// CreateConnectToken gera um token temporário para inicializar o Pluggy Connect Widget
func (c *Client) CreateConnectToken(ctx context.Context, clientUserID string) (string, error) {
	payload := map[string]any{
		"options": map[string]string{
			"clientUserId": clientUserID,
		},
	}
	var res ConnectTokenResponse
	if err := c.doRequest(ctx, http.MethodPost, "/connect_token", payload, &res); err != nil {
		return "", err
	}
	return res.AccessToken, nil
}

// GetItem busca os detalhes de uma conexão
func (c *Client) GetItem(ctx context.Context, itemID string) (*Item, error) {
	var item Item
	if err := c.doRequest(ctx, http.MethodGet, "/items/"+url.PathEscape(itemID), nil, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// SyncItem solicita ao Pluggy a sincronização atualizada de dados bancários
func (c *Client) SyncItem(ctx context.Context, itemID string) (*Item, error) {
	var item Item
	if err := c.doRequest(ctx, http.MethodPatch, "/items/"+url.PathEscape(itemID), map[string]any{}, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// DeleteItem desconecta e exclui o item no Pluggy
func (c *Client) DeleteItem(ctx context.Context, itemID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/items/"+url.PathEscape(itemID), nil, nil)
}

// GetAccounts busca todas as contas associadas a um Item (corrente, poupança, cartão de crédito)
func (c *Client) GetAccounts(ctx context.Context, itemID string) ([]Account, error) {
	path := fmt.Sprintf("/accounts?itemId=%s", url.QueryEscape(itemID))
	var res AccountsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}
	return res.Results, nil
}

// GetTransactions busca todas as transações de uma conta com suporte a paginação
func (c *Client) GetTransactions(ctx context.Context, accountID string, from, to string) ([]Transaction, error) {
	var allTransactions []Transaction
	page := 1
	pageSize := 500

	for {
		params := url.Values{}
		params.Set("accountId", accountID)
		params.Set("pageSize", fmt.Sprintf("%d", pageSize))
		params.Set("page", fmt.Sprintf("%d", page))
		if from != "" {
			params.Set("from", from)
		}
		if to != "" {
			params.Set("to", to)
		}

		path := "/transactions?" + params.Encode()
		var res TransactionsResponse
		if err := c.doRequest(ctx, http.MethodGet, path, nil, &res); err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, res.Results...)

		if res.Page >= res.TotalPages || len(res.Results) == 0 {
			break
		}
		page++
	}

	return allTransactions, nil
}

// GetBills busca as faturas de uma conta de cartão de crédito
func (c *Client) GetBills(ctx context.Context, accountID string) ([]CreditCardBill, error) {
	path := fmt.Sprintf("/bills?accountId=%s", url.QueryEscape(accountID))
	var res BillsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &res); err != nil {
		// Algumas contas de cartão podem ainda não ter faturas geradas, não bloqueia
		return nil, nil
	}
	return res.Results, nil
}

// GetInvestments busca a carteira de investimentos do item
func (c *Client) GetInvestments(ctx context.Context, itemID string) ([]Investment, error) {
	path := fmt.Sprintf("/investments?itemId=%s", url.QueryEscape(itemID))
	var res InvestmentsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, err
	}
	return res.Results, nil
}

// GetInvestmentTransactions busca o histórico de movimentações de um investimento
func (c *Client) GetInvestmentTransactions(ctx context.Context, investmentID string) ([]InvestmentTransaction, error) {
	path := fmt.Sprintf("/investments/%s/transactions", url.PathEscape(investmentID))
	var res InvestmentTransactionsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, nil
	}
	return res.Results, nil
}

// ListWebhooks lista os webhooks registrados na aplicação Pluggy
func (c *Client) ListWebhooks(ctx context.Context) ([]WebhookRegistration, error) {
	var res WebhooksListResponse
	if err := c.doRequest(ctx, http.MethodGet, "/webhooks", nil, &res); err != nil {
		return nil, err
	}
	return res.Results, nil
}

// CreateWebhook registra um novo webhook para receber eventos
func (c *Client) CreateWebhook(ctx context.Context, webhookURL, event string) (*WebhookRegistration, error) {
	if event == "" {
		event = "all"
	}
	payload := WebhookRegistration{
		Event: event,
		URL:   webhookURL,
	}
	var res WebhookRegistration
	if err := c.doRequest(ctx, http.MethodPost, "/webhooks", payload, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// DeleteWebhook remove um webhook registrado
func (c *Client) DeleteWebhook(ctx context.Context, webhookID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/webhooks/"+url.PathEscape(webhookID), nil, nil)
}

