package santander

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"accounting/internal/adapter/config"
	"accounting/internal/domain"
	"accounting/internal/port"
)

// Client implements port.BankStatementProvider for Santander Open API - Contas Correntes
type Client struct {
	cfg        config.SantanderConfig
	logger     port.Logger
	httpClient *http.Client

	// Token cache per client_id
	tokenMu  sync.Mutex
	tokenMap map[string]*tokenEntry
}

type tokenEntry struct {
	AccessToken string
	ExpiresAt   time.Time
}

var _ port.BankStatementProvider = (*Client)(nil)

// NewClient initializes Santander API client with mTLS configuration if provided
func NewClient(cfg config.SantanderConfig, logger port.Logger) (*Client, error) {
	client := &Client{
		cfg:      cfg,
		logger:   logger,
		tokenMap: make(map[string]*tokenEntry),
	}

	tlsConfig, err := client.buildTLSConfig()
	if err != nil {
		if logger != nil {
			logger.IPrintf(1, "Warning: Santander mTLS setup could not be completed (%v). Fallback to standard TLS.", err)
		}
		tlsConfig = &tls.Config{InsecureSkipVerify: false}
	} else if len(tlsConfig.Certificates) > 0 {
		if logger != nil {
			logger.IPrintf(0, "Santander mTLS setup completed successfully with certificate: %s", client.cfg.CertificatePath)
		}
	} else {
		if logger != nil {
			logger.IPrintf(1, "Warning: No client certificate found for Santander mTLS. Outgoing requests will lack client credentials.")
		}
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client.httpClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}

func (c *Client) ProviderName() string {
	return "santander_api"
}

// buildTLSConfig configures client certificate from .pfx / .crt if available
func (c *Client) buildTLSConfig() (*tls.Config, error) {
	certPath := c.cfg.CertificatePath
	if certPath == "" {
		if _, err := os.Stat("certs/26534905.pfx"); err == nil {
			certPath = "certs/26534905.pfx"
		} else if _, err := os.Stat("certs/certificado.crt"); err == nil {
			certPath = "certs/certificado.crt"
		}
	}

	if certPath == "" {
		return &tls.Config{}, nil
	}

	tlsConfig := &tls.Config{}

	if strings.HasSuffix(strings.ToLower(certPath), ".pfx") {
		pfxData, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("reading pfx cert from %s: %w", certPath, err)
		}

		password := c.cfg.CertPassword
		if password == "" {
			password = os.Getenv("SANTANDER_CERT_PASSWORD")
		}

		privKey, cert, caCerts, err := pkcs12.DecodeChain(pfxData, password)
		if err != nil {
			return nil, fmt.Errorf("decoding pfx cert: %w", err)
		}

		rsaKey, ok := privKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key in pfx is not RSA")
		}

		tlsCert := tls.Certificate{
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  rsaKey,
			Leaf:        cert,
		}
		for _, ca := range caCerts {
			tlsCert.Certificate = append(tlsCert.Certificate, ca.Raw)
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	} else if strings.HasSuffix(strings.ToLower(certPath), ".crt") {
		certBytes, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("reading crt from %s: %w", certPath, err)
		}
		caPool, _ := x509.SystemCertPool()
		if caPool == nil {
			caPool = x509.NewCertPool()
		}
		caPool.AppendCertsFromPEM(certBytes)
		tlsConfig.RootCAs = caPool
	}

	return tlsConfig, nil
}

// getBaseURL returns base URL depending on environment
func (c *Client) getBaseURL() string {
	if strings.EqualFold(c.cfg.Environment, "production") || strings.EqualFold(c.cfg.Environment, "prod") {
		return c.cfg.ProdBaseURL
	}
	return c.cfg.SandboxBaseURL
}

func sanitizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// formatAccountBranchConta formats agency and account into "AGENCIA.CONTA" format required by Santander
// Agência must have 4 digits (e.g. 0985), Conta must have 12 digits (e.g. 000130010014)
func (c *Client) formatAccountBranchConta(branch, number string) string {
	cleanBranch := sanitizeDigits(branch)
	if len(cleanBranch) < 4 {
		cleanBranch = strings.Repeat("0", 4-len(cleanBranch)) + cleanBranch
	} else if len(cleanBranch) > 4 {
		cleanBranch = cleanBranch[len(cleanBranch)-4:]
	}

	cleanNumber := sanitizeDigits(number)
	if len(cleanNumber) < 12 {
		cleanNumber = strings.Repeat("0", 12-len(cleanNumber)) + cleanNumber
	} else if len(cleanNumber) > 12 {
		cleanNumber = cleanNumber[len(cleanNumber)-12:]
	}

	return fmt.Sprintf("%s.%s", cleanBranch, cleanNumber)
}

// getAccessToken authenticates with Santander OAuth2 endpoint
func (c *Client) getAccessToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if entry, exists := c.tokenMap[clientID]; exists {
		if time.Now().Before(entry.ExpiresAt.Add(-1 * time.Minute)) {
			return entry.AccessToken, nil
		}
	}

	tokenURL := c.cfg.TokenURL
	if tokenURL == "" {
		tokenURL = fmt.Sprintf("%s/auth/oauth/v2/token", c.getBaseURL())
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("client_id", clientID)
	req.Header.Set("client_secret", clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token request error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if tokenResp.ExpiresIn <= 0 {
		expiresAt = time.Now().Add(1 * time.Hour)
	}

	c.tokenMap[clientID] = &tokenEntry{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   expiresAt,
	}

	return tokenResp.AccessToken, nil
}

// FetchStatements retrieves transactions from Santander Open API - Contas Correntes
func (c *Client) FetchStatements(ctx context.Context, account *domain.Account, creds *domain.AccountSantanderAPI, initialDate, finalDate time.Time) ([]port.BankTransaction, error) {
	if account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	baseURL := c.getBaseURL()
	bankID := c.cfg.BankID
	if bankID == "" {
		bankID = "90400888000142"
	}

	branchConta := c.formatAccountBranchConta(account.AccountBranch, account.AccountNumber)
	initDateStr := initialDate.Format("2006-01-02")
	finalDateStr := finalDate.Format("2006-01-02")

	var token string
	if creds != nil && creds.ClientID != "" && creds.ClientSecret != "" {
		t, err := c.getAccessToken(ctx, creds.ClientID, creds.ClientSecret)
		if err != nil {
			if c.logger != nil {
				c.logger.IPrintf(1, "OAuth token warning for account %d: %v. Continuing request with mTLS.", account.ID, err)
			}
		} else {
			token = t
		}
	}

	var allTransactions []port.BankTransaction
	offset := 1
	limit := 50

	for {
		reqURL := fmt.Sprintf("%s/bank_account_information/v1/banks/%s/statements/%s?_offset=%d&_limit=%d&initialDate=%s&finalDate=%s",
			baseURL, bankID, branchConta, offset, limit, initDateStr, finalDateStr)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating statement request: %w", err)
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if creds != nil && creds.ClientID != "" {
			req.Header.Set("X-Application-Key", creds.ClientID)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "application/json")

		if c.logger != nil {
			c.logger.IPrintf(0, "Santander API Statement Request: GET %s (X-Application-Key: %s)", reqURL, creds.ClientID)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching statements from %s: %w", reqURL, err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if c.logger != nil {
			c.logger.IPrintf(0, "Santander API Statement Response (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("santander statement error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
		}

		items, hasMore, err := c.parseStatementResponse(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("parsing santander statement response: %w", err)
		}

		if c.logger != nil {
			c.logger.IPrintf(0, "Santander API Parsed %d transaction(s) from response (offset: %d, hasMore: %t)", len(items), offset, hasMore)
		}

		allTransactions = append(allTransactions, items...)

		if !hasMore || len(items) == 0 {
			break
		}
		offset += len(items)
	}

	return allTransactions, nil
}

// parseStatementResponse parses JSON response returned by Santander Statement API
func (c *Client) parseStatementResponse(data []byte) ([]port.BankTransaction, bool, error) {
	var raw struct {
		Transactions      []santanderTransactionRaw `json:"transactions"`
		Statements        []santanderTransactionRaw `json:"statements"`
		Items             []santanderTransactionRaw `json:"items"`
		Data              []santanderTransactionRaw `json:"data"`
		Content           []santanderTransactionRaw `json:"content"`
		UnderscoreContent []santanderTransactionRaw `json:"_content"`
		Pagination        struct {
			TotalElements int  `json:"totalElements"`
			HasMore       bool `json:"hasMore"`
			Next          bool `json:"next"`
		} `json:"pagination"`
		Pageable struct {
			MoreElements bool        `json:"_moreElements"`
			TotalPages   interface{} `json:"totalPages"`
			TotalRecords interface{} `json:"totalRecords"`
		} `json:"_pageable"`
		HasMore bool `json:"hasMore"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		var list []santanderTransactionRaw
		if errList := json.Unmarshal(data, &list); errList == nil {
			return convertRawList(list), false, nil
		}
		return nil, false, err
	}

	txList := raw.UnderscoreContent
	if len(txList) == 0 {
		txList = raw.Content
	}
	if len(txList) == 0 {
		txList = raw.Transactions
	}
	if len(txList) == 0 {
		txList = raw.Statements
	}
	if len(txList) == 0 {
		txList = raw.Items
	}
	if len(txList) == 0 {
		txList = raw.Data
	}

	result := convertRawList(txList)
	hasMore := raw.HasMore || raw.Pagination.HasMore || raw.Pagination.Next || raw.Pageable.MoreElements
	return result, hasMore, nil
}

type santanderTransactionRaw struct {
	CreditDebitType    string      `json:"creditDebitType"`
	TransactionName    string      `json:"transactionName"`
	HistoricComplement string      `json:"historicComplement"`
	Amount             interface{} `json:"amount"`
	TransactionDate    string      `json:"transactionDate"`
	DocumentNumber     string      `json:"documentNumber"`
	DocumentType       string      `json:"documentType"`
}

func convertRawList(rawList []santanderTransactionRaw) []port.BankTransaction {
	var list []port.BankTransaction
	for _, item := range rawList {
		cType := strings.ToUpper(strings.TrimSpace(item.CreditDebitType))
		name := strings.TrimSpace(item.TransactionName)
		comp := strings.TrimSpace(item.HistoricComplement)
		amt := math.Abs(parseAmount(item.Amount))
		tDate := parseDate(item.TransactionDate)
		doc := strings.TrimSpace(item.DocumentNumber)

		list = append(list, port.BankTransaction{
			CreditDebitType:    cType,
			TransactionName:    name,
			HistoricComplement: comp,
			Amount:             amt,
			TransactionDate:    tDate,
			DocumentNumber:     doc,
		})
	}
	return list
}

func parseAmount(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		clean := strings.ReplaceAll(v, ",", ".")
		f, _ := strconv.ParseFloat(clean, 64)
		return f
	default:
		return 0
	}
}

func parseDate(val string) time.Time {
	val = strings.TrimSpace(val)
	formats := []string{
		"02/01/2006",
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t
		}
	}
	return time.Now()
}
