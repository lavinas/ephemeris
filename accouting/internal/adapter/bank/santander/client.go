package santander

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
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

// formatAccountBranchConta formats agency and account into "AGENCIA.CONTA" format required by Santander
// e.g. branch: "0985", number: "13001001-4" -> "0985.000013001001" or sanitized
func (c *Client) formatAccountBranchConta(branch, number string) string {
	cleanBranch := strings.TrimSpace(branch)
	// Santander expects digits only, without hyphen
	cleanNumber := strings.ReplaceAll(number, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ".", "")
	cleanNumber = strings.TrimSpace(cleanNumber)

	// In sandbox collection: "0000.000011112222"
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
		tokenURL = fmt.Sprintf("%s/oauth/v1/token", c.getBaseURL())
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

	var allTransactions []port.BankTransaction
	offset := 1
	limit := 50

	var token string
	if creds != nil && creds.ClientID != "" && creds.ClientSecret != "" {
		t, err := c.getAccessToken(ctx, creds.ClientID, creds.ClientSecret)
		if err != nil {
			if c.logger != nil {
				c.logger.IPrintf(1, "OAuth token warning for account %d: %v. Continuing request.", account.ID, err)
			}
		} else {
			token = t
		}
	}

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

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching statements from %s: %w", reqURL, err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("santander statement error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
		}

		items, hasMore, err := c.parseStatementResponse(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("parsing santander statement response: %w", err)
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
		Transactions []santanderTransactionRaw `json:"transactions"`
		Statements   []santanderTransactionRaw `json:"statements"`
		Items        []santanderTransactionRaw `json:"items"`
		Data         []santanderTransactionRaw `json:"data"`
		Pagination   struct {
			TotalElements int  `json:"totalElements"`
			HasMore       bool `json:"hasMore"`
			Next          bool `json:"next"`
		} `json:"pagination"`
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

	result := convertRawList(txList)
	hasMore := raw.HasMore || raw.Pagination.HasMore || raw.Pagination.Next
	return result, hasMore, nil
}

type santanderTransactionRaw struct {
	TransactionID       string      `json:"transactionId"`
	ID                  string      `json:"id"`
	TransactionStatus   string      `json:"transactionStatus"`
	Status              string      `json:"status"`
	TransactionSign     string      `json:"transactionSign"` // "+" or "-"
	Sign                string      `json:"sign"`
	TransactionName     string      `json:"transactionName"`
	Description         string      `json:"description"`
	Memo                string      `json:"memo"`
	TransactionType     string      `json:"transactionType"`
	Type                string      `json:"type"`
	TransactionAmount   interface{} `json:"transactionAmount"`
	Amount              interface{} `json:"amount"`
	TransactionCurrency string      `json:"transactionCurrency"`
	Currency            string      `json:"currency"`
	TransactionDate     string      `json:"transactionDate"`
	Date                string      `json:"date"`
	PartieDocument      string      `json:"partieDocument"`
	PartieType          string      `json:"partieType"`
	PartieBankCode      string      `json:"partieBankCode"`
	PartieBranchCode    string      `json:"partieBranchCode"`
	PartieAccount       string      `json:"partieAccount"`
	CounterPartie       struct {
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

		name := item.TransactionName
		if name == "" {
			name = item.Description
		}
		if name == "" {
			name = item.Memo
		}

		status := item.TransactionStatus
		if status == "" {
			status = item.Status
		}

		tType := item.TransactionType
		if tType == "" {
			tType = item.Type
		}

		amt := parseAmount(item.TransactionAmount)
		if amt == 0 {
			amt = parseAmount(item.Amount)
		}

		curr := item.TransactionCurrency
		if curr == "" {
			curr = item.Currency
		}
		if curr == "" {
			curr = "BRL"
		}

		dateStr := item.TransactionDate
		if dateStr == "" {
			dateStr = item.Date
		}
		tDate := parseDate(dateStr)

		pDoc := item.PartieDocument
		if pDoc == "" {
			pDoc = item.CounterPartie.Document
		}

		pType := item.PartieType
		if pType == "" {
			pType = item.CounterPartie.Type
		}

		pBank := item.PartieBankCode
		if pBank == "" {
			pBank = item.CounterPartie.BankCode
		}

		pBranch := item.PartieBranchCode
		if pBranch == "" {
			pBranch = item.CounterPartie.BranchCode
		}

		pAcc := item.PartieAccount
		if pAcc == "" {
			pAcc = item.CounterPartie.Account
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

