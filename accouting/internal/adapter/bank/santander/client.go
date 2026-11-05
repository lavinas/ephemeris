package santander

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
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

// Client implements port.BankStatementProvider for Santander API
type Client struct {
	cfg        config.SantanderConfig
	logger     port.Logger
	httpClient *http.Client

	// Token cache per client_id
	tokenMu   sync.Mutex
	tokenMap  map[string]*tokenEntry
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
		// Look for default cert files in ./certs
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
		// CRT files
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
	if strings.ToLower(c.cfg.Environment) == "production" || strings.ToLower(c.cfg.Environment) == "prod" {
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

// getAccessToken authenticates with Santander OAuth2 if needed
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
	req.Header.Set("client_id", clientID)
	req.Header.Set("client_secret", clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Em caso de falha de rede/mock em sandbox sem backend real levantado
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

// FetchStatements retrieves transactions from Santander API
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

	// Date format YYYY-MM-DD
	initDateStr := initialDate.Format("2006-01-02")
	finalDateStr := finalDate.Format("2006-01-02")

	var token string
	if creds != nil && creds.ClientID != "" && creds.ClientSecret != "" {
		t, err := c.getAccessToken(ctx, creds.ClientID, creds.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("oauth token error for account %d: %w", account.ID, err)
		}
		token = t
	}

	// Try Open Finance detailed statements route first if accountId is available or discoverable
	cleanBranch := sanitizeDigits(account.AccountBranch)
	if len(cleanBranch) < 4 {
		cleanBranch = strings.Repeat("0", 4-len(cleanBranch)) + cleanBranch
	}
	cleanNumber := sanitizeDigits(account.AccountNumber)
	if len(cleanNumber) < 12 {
		cleanNumber = strings.Repeat("0", 12-len(cleanNumber)) + cleanNumber
	}

	var accountIDParam string
	if creds != nil && creds.ClientID != "" {
		accountIDParam = c.discoverAccountID(ctx, token, creds.ClientID, bankID, cleanBranch, cleanNumber)
	}

	if accountIDParam != "" {
		txs, err := c.fetchStatementsWithQueryParams(ctx, token, creds, bankID, cleanBranch, cleanNumber, accountIDParam, initDateStr, finalDateStr)
		if err == nil && len(txs) > 0 {
			if c.logger != nil {
				c.logger.IPrintf(0, "Successfully fetched %d statement(s) via Open Finance detailed endpoint", len(txs))
			}
			return txs, nil
		}
		if c.logger != nil {
			c.logger.IPrintf(0, "Open Finance detailed endpoint did not return transactions (%v), proceeding with corporate path endpoint", err)
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

// discoverAccountID queries /bank_account_information/v1/banks/{bank_id}/accounts to find Open Finance accountId
func (c *Client) discoverAccountID(ctx context.Context, token, clientID, bankID, branch, number string) string {
	baseURL := c.getBaseURL()
	reqURL := fmt.Sprintf("%s/bank_account_information/v1/banks/%s/accounts?_offset=1&_limit=50", baseURL, bankID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ""
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if clientID != "" {
		req.Header.Set("X-Application-Key", clientID)
	}
	req.Header.Set("Accept", "application/json")

	if c.logger != nil {
		c.logger.IPrintf(0, "Santander API Discover Accounts Request: GET %s", reqURL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if c.logger != nil {
			c.logger.IPrintf(1, "Santander API Discover Accounts failed: %v", err)
		}
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	if c.logger != nil {
		c.logger.IPrintf(0, "Santander API Discover Accounts Response (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	var raw struct {
		Accounts          []santanderAccountItem `json:"accounts"`
		Content           []santanderAccountItem `json:"content"`
		UnderscoreContent []santanderAccountItem `json:"_content"`
		Data              []santanderAccountItem `json:"data"`
		Items             []santanderAccountItem `json:"items"`
	}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		var list []santanderAccountItem
		if errList := json.Unmarshal(bodyBytes, &list); errList == nil {
			return findAccountID(list, branch, number)
		}
		return ""
	}

	items := raw.Accounts
	if len(items) == 0 {
		items = raw.Content
	}
	if len(items) == 0 {
		items = raw.UnderscoreContent
	}
	if len(items) == 0 {
		items = raw.Data
	}
	if len(items) == 0 {
		items = raw.Items
	}

	return findAccountID(items, branch, number)
}

type santanderAccountItem struct {
	AccountID     string `json:"accountId"`
	ID            string `json:"id"`
	BranchCode    string `json:"branchCode"`
	Agency        string `json:"agency"`
	AccountNumber string `json:"accountNumber"`
	Number        string `json:"number"`
}

func findAccountID(items []santanderAccountItem, branch, number string) string {
	cleanBranch := sanitizeDigits(branch)
	cleanNumber := sanitizeDigits(number)

	for _, item := range items {
		itemBranch := sanitizeDigits(item.BranchCode)
		if itemBranch == "" {
			itemBranch = sanitizeDigits(item.Agency)
		}
		itemNumber := sanitizeDigits(item.AccountNumber)
		if itemNumber == "" {
			itemNumber = sanitizeDigits(item.Number)
		}

		if (itemBranch == "" || strings.HasSuffix(cleanBranch, itemBranch) || strings.HasSuffix(itemBranch, cleanBranch)) &&
			(itemNumber == "" || strings.HasSuffix(cleanNumber, itemNumber) || strings.HasSuffix(itemNumber, cleanNumber)) {
			if item.AccountID != "" {
				return item.AccountID
			}
			if item.ID != "" {
				return item.ID
			}
		}
	}

	if len(items) == 1 {
		if items[0].AccountID != "" {
			return items[0].AccountID
		}
		return items[0].ID
	}
	return ""
}

// fetchStatementsWithQueryParams queries Open Finance statements using query params
func (c *Client) fetchStatementsWithQueryParams(ctx context.Context, token string, creds *domain.AccountSantanderAPI, bankID, branch, number, accountID, initDateStr, finalDateStr string) ([]port.BankTransaction, error) {
	baseURL := c.getBaseURL()
	var allTransactions []port.BankTransaction
	offset := 1
	limit := 50

	for {
		reqURL := fmt.Sprintf("%s/bank_account_information/v1/banks/%s/statements?branchCode=%s&accountNumber=%s&accountId=%s&initialDate=%s&finalDate=%s&_offset=%d&_limit=%d",
			baseURL, bankID, branch, number, accountID, initDateStr, finalDateStr, offset, limit)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if creds != nil && creds.ClientID != "" {
			req.Header.Set("X-Application-Key", creds.ClientID)
		}
		req.Header.Set("Accept", "application/json")

		if c.logger != nil {
			c.logger.IPrintf(0, "Santander Open Finance Statement Request: GET %s", reqURL)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if c.logger != nil {
			c.logger.IPrintf(0, "Santander Open Finance Statement Response (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}

		items, hasMore, err := c.parseStatementResponse(bodyBytes)
		if err != nil {
			return nil, err
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
	// Flexible structure handling Santander API schema
	var raw struct {
		Transactions      []santanderTransactionRaw `json:"transactions"`
		Statements        []santanderTransactionRaw `json:"statements"`
		Items             []santanderTransactionRaw `json:"items"`
		Data              []santanderTransactionRaw `json:"data"`
		Content           []santanderTransactionRaw `json:"content"`
		UnderscoreContent []santanderTransactionRaw `json:"_content"`
		Lancamentos       []santanderTransactionRaw `json:"lancamentos"`
		Transacoes        []santanderTransactionRaw `json:"transacoes"`
		Pagination        struct {
			TotalElements int  `json:"totalElements"`
			HasMore       bool `json:"hasMore"`
			Next          bool `json:"next"`
		} `json:"pagination"`
		Pageable struct {
			MoreElements    bool        `json:"_moreElements"`
			MoreElementsAlt bool        `json:"moreElements"`
			TotalPages      interface{} `json:"totalPages"`
			TotalRecords    interface{} `json:"totalRecords"`
		} `json:"_pageable"`
		PageableAlt struct {
			MoreElements    bool `json:"_moreElements"`
			MoreElementsAlt bool `json:"moreElements"`
		} `json:"pageable"`
		HasMore bool `json:"hasMore"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		// Could also be an array directly
		var list []santanderTransactionRaw
		if errList := json.Unmarshal(data, &list); errList == nil {
			txs := convertRawList(list)
			return txs, false, nil
		}
		return nil, false, err
	}

	txList := raw.Transactions
	if len(txList) == 0 {
		txList = raw.Statements
	}
	if len(txList) == 0 {
		txList = raw.Items
	}
	if len(txList) == 0 {
		txList = raw.Data
	}
	if len(txList) == 0 {
		txList = raw.Content
	}
	if len(txList) == 0 {
		txList = raw.UnderscoreContent
	}
	if len(txList) == 0 {
		txList = raw.Lancamentos
	}
	if len(txList) == 0 {
		txList = raw.Transacoes
	}

	result := convertRawList(txList)
	hasMore := raw.HasMore || raw.Pagination.HasMore || raw.Pagination.Next ||
		raw.Pageable.MoreElements || raw.Pageable.MoreElementsAlt ||
		raw.PageableAlt.MoreElements || raw.PageableAlt.MoreElementsAlt
	return result, hasMore, nil
}

type santanderTransactionRaw struct {
	TransactionID                  string      `json:"transactionId"`
	ID                             string      `json:"id"`
	TransactionStatus              string      `json:"transactionStatus"`
	Status                         string      `json:"status"`
	CompletedAuthorisedPaymentType string      `json:"completedAuthorisedPaymentType"`
	TransactionSign                string      `json:"transactionSign"` // "+" or "-"
	Sign                           string      `json:"sign"`
	Sinal                          string      `json:"sinal"`
	CreditDebitIndicator           string      `json:"creditDebitIndicator"`
	CreditDebitType                string      `json:"creditDebitType"` // "DEBITO" or "CREDITO"
	TransactionName                string      `json:"transactionName"`
	Description                    string      `json:"description"`
	Descricao                      string      `json:"descricao"`
	HistoricComplement             string      `json:"historicComplement"`
	Historico                      string      `json:"historico"`
	Memo                           string      `json:"memo"`
	TransactionType                string      `json:"transactionType"`
	Type                           string      `json:"type"`
	Tipo                           string      `json:"tipo"`
	TransactionAmount              interface{} `json:"transactionAmount"`
	Amount                         interface{} `json:"amount"`
	Valor                          interface{} `json:"valor"`
	Value                          interface{} `json:"value"`
	TransactionCurrency            string      `json:"transactionCurrency"`
	Currency                       string      `json:"currency"`
	Moeda                          string      `json:"moeda"`
	TransactionDate                string      `json:"transactionDate"`
	Date                           string      `json:"date"`
	Data                           string      `json:"data"`
	BookingDate                    string      `json:"bookingDate"`
	ValueDate                      string      `json:"valueDate"`
	PartieDocument                 string      `json:"partieDocument"`
	PartieCnpjCpf                  string      `json:"partieCnpjCpf"`
	PartieType                     string      `json:"partieType"`
	PartiePersonType               string      `json:"partiePersonType"`
	PartieBankCode                 string      `json:"partieBankCode"`
	PartieCompeCode                string      `json:"partieCompeCode"`
	PartieBranchCode               string      `json:"partieBranchCode"`
	PartieAccount                  string      `json:"partieAccount"`
	PartieNumber                   string      `json:"partieNumber"`
	PartieCheckDigit               string      `json:"partieCheckDigit"`
	CounterPartie                  struct {
		Document   string `json:"document"`
		Type       string `json:"type"`
		BankCode   string `json:"bankCode"`
		BranchCode string `json:"branchCode"`
		Account    string `json:"account"`
	} `json:"counterPartie"`
}

func convertRawList(rawList []santanderTransactionRaw) []port.BankTransaction {
	var list []port.BankTransaction
	for _, item := range rawList {
		tID := item.TransactionID
		if tID == "" {
			tID = item.ID
		}

		sign := item.TransactionSign
		if sign == "" {
			sign = item.Sign
		}
		if sign == "" {
			sign = item.Sinal
		}
		if sign == "" {
			if strings.EqualFold(item.CreditDebitIndicator, "CRDT") || strings.EqualFold(item.CreditDebitIndicator, "CREDIT") || strings.EqualFold(item.CreditDebitIndicator, "C") {
				sign = "+"
			} else if strings.EqualFold(item.CreditDebitIndicator, "DBIT") || strings.EqualFold(item.CreditDebitIndicator, "DEBIT") || strings.EqualFold(item.CreditDebitIndicator, "D") {
				sign = "-"
			}
		}
		if sign == "" {
			if strings.EqualFold(item.CreditDebitType, "DEBITO") || strings.EqualFold(item.CreditDebitType, "DEBIT") || strings.EqualFold(item.CreditDebitType, "D") {
				sign = "-"
			} else if strings.EqualFold(item.CreditDebitType, "CREDITO") || strings.EqualFold(item.CreditDebitType, "CREDIT") || strings.EqualFold(item.CreditDebitType, "C") {
				sign = "+"
			}
		}

		name := item.TransactionName
		if name == "" {
			name = item.Description
		}
		if name == "" {
			name = item.Descricao
		}
		if name == "" {
			name = item.Historico
		}
		if name == "" {
			name = item.Memo
		}

		status := item.TransactionStatus
		if status == "" {
			status = item.Status
		}
		if status == "" {
			status = item.CompletedAuthorisedPaymentType
		}

		tType := item.TransactionType
		if tType == "" {
			tType = item.Type
		}
		if tType == "" {
			tType = item.Tipo
		}
		if tType == "" && item.CreditDebitType != "" {
			tType = item.CreditDebitType
		}

		amt := parseAmount(item.TransactionAmount)
		if amt == 0 {
			amt = parseAmount(item.Amount)
		}
		if amt == 0 {
			amt = parseAmount(item.Valor)
		}
		if amt == 0 {
			amt = parseAmount(item.Value)
		}
		if amt < 0 {
			sign = "-"
			amt = math.Abs(amt)
		}

		curr := item.TransactionCurrency
		if curr == "" {
			curr = item.Currency
		}
		if curr == "" {
			curr = item.Moeda
		}
		if curr == "" {
			curr = "BRL"
		}

		dateStr := item.TransactionDate
		if dateStr == "" {
			dateStr = item.Date
		}
		if dateStr == "" {
			dateStr = item.Data
		}
		if dateStr == "" {
			dateStr = item.BookingDate
		}
		if dateStr == "" {
			dateStr = item.ValueDate
		}
		tDate := parseDate(dateStr)

		pDoc := item.PartieDocument
		if pDoc == "" {
			pDoc = item.PartieCnpjCpf
		}
		if pDoc == "" {
			pDoc = item.HistoricComplement
		}
		if pDoc == "" {
			pDoc = item.CounterPartie.Document
		}

		pType := item.PartieType
		if pType == "" {
			pType = item.PartiePersonType
		}
		if pType == "" {
			pType = item.CounterPartie.Type
		}

		pBank := item.PartieBankCode
		if pBank == "" {
			pBank = item.PartieCompeCode
		}
		if pBank == "" {
			pBank = item.CounterPartie.BankCode
		}

		pBranch := item.PartieBranchCode
		if pBranch == "" {
			pBranch = item.CounterPartie.BranchCode
		}

		pAcc := item.PartieAccount
		if pAcc == "" {
			if item.PartieNumber != "" {
				pAcc = item.PartieNumber
				if item.PartieCheckDigit != "" {
					pAcc = fmt.Sprintf("%s-%s", item.PartieNumber, item.PartieCheckDigit)
				}
			}
		}
		if pAcc == "" {
			pAcc = item.CounterPartie.Account
		}

		if tID == "" {
			// Generate deterministic synthetic ID when API doesn't provide transactionId
			rawHash := fmt.Sprintf("%s|%s|%.2f|%s|%s", dateStr, sign, amt, name, pDoc)
			h := sha256.Sum256([]byte(rawHash))
			tID = fmt.Sprintf("santander_%x", h[:16])
		}

		list = append(list, port.BankTransaction{
			TransactionID:       tID,
			TransactionStatus:   status,
			TransactionSign:     sign,
			TransactionName:     name,
			TransactionType:     tType,
			TransactionAmount:   amt,
			TransactionCurrency: curr,
			TransactionDate:     tDate,
			PartieDocument:      pDoc,
			PartieType:          pType,
			PartieBankCode:      pBank,
			PartieBranchCode:    pBranch,
			PartieAccount:       pAcc,
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
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"02/01/2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t
		}
	}
	return time.Now()
}

