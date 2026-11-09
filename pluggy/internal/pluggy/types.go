package pluggy

import "time"

// AuthResponse representa a resposta do endpoint /auth
type AuthResponse struct {
	APIKey string `json:"apiKey"`
}

// ConnectTokenResponse representa a resposta de criação de connect token
type ConnectTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

// Connector representa os metadados do banco/instituição financeira
type Connector struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	PrimaryColor string `json:"primaryColor"`
	InstitutionURL string `json:"institutionUrl"`
	Country      string `json:"country"`
	Type         string `json:"type"`
	ImageURL     string `json:"imageUrl"`
	HasMFA       bool   `json:"hasMFA"`
}

// ItemError representa eventuais erros retornados na sincronização do item
type ItemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Item representa a conexão de uma instituição no Pluggy
type Item struct {
	ID              string     `json:"id"`
	Connector       Connector  `json:"connector"`
	Status          string     `json:"status"`          // UPDATING, UPDATED, WAITING_USER_INPUT, LOGIN_ERROR, OUTDATED
	ExecutionStatus string     `json:"executionStatus"` // SUCCESS, PARTIAL_SUCCESS, CREATED, ERROR
	ClientUserID    string     `json:"clientUserId"`
	Error           *ItemError `json:"error"`
	LastUpdatedAt   *time.Time `json:"lastUpdatedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// BankData representa detalhes adicionais para contas bancárias
type BankData struct {
	TransferNumber              string   `json:"transferNumber"`
	ClosingBalance              *float64 `json:"closingBalance"`
	AutomaticallyInvestedBalance *float64 `json:"automaticallyInvestedBalance"`
	OverdraftContractedLimit    *float64 `json:"overdraftContractedLimit"`
	OverdraftUsedLimit          *float64 `json:"overdraftUsedLimit"`
	UnarrangedOverdraftAmount   *float64 `json:"unarrangedOverdraftAmount"`
	HasReservedBalance          *bool    `json:"hasReservedBalance"`
}

// CreditData representa detalhes adicionais para cartões de crédito
type CreditData struct {
	Level                string   `json:"level"`
	Brand                string   `json:"brand"`
	BalanceCloseDate     string   `json:"balanceCloseDate"`
	BalanceDueDate       string   `json:"balanceDueDate"`
	AvailableCreditLimit *float64 `json:"availableCreditLimit"`
	MinimumPayment       *float64 `json:"minimumPayment"`
	CreditLimit          *float64 `json:"creditLimit"`
	Status               string   `json:"status"`
	HolderType           string   `json:"holderType"`
}

// Account representa uma conta corrente, poupança ou cartão de crédito
type Account struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"`    // BANK, CREDIT
	Subtype       string      `json:"subtype"` // CHECKING_ACCOUNT, SAVINGS_ACCOUNT, CREDIT_CARD
	Number        string      `json:"number"`
	Name          string      `json:"name"`
	MarketingName string      `json:"marketingName"`
	Balance       float64     `json:"balance"`
	CurrencyCode  string      `json:"currencyCode"`
	ItemID        string      `json:"itemId"`
	TaxNumber     string      `json:"taxNumber"`
	Owner         string      `json:"owner"`
	BankData      *BankData   `json:"bankData"`
	CreditData    *CreditData `json:"creditData"`
}

type AccountsResponse struct {
	Total      int       `json:"total"`
	TotalPages int       `json:"totalPages"`
	Page       int       `json:"page"`
	Results    []Account `json:"results"`
}

// MerchantData representa os dados enriquecidos de estabelecimento
type MerchantData struct {
	Name         string `json:"name"`
	BusinessName string `json:"businessName"`
	CNPJ         string `json:"cnpj"`
	CNAE         string `json:"cnae"`
	Category     string `json:"category"`
}

// Transaction representa uma movimentação no extrato ou fatura
type Transaction struct {
	ID             string        `json:"id"`
	Description    string        `json:"description"`
	DescriptionRaw string        `json:"descriptionRaw"`
	CurrencyCode   string        `json:"currencyCode"`
	Amount         float64       `json:"amount"`
	Date           time.Time     `json:"date"`
	Balance        *float64      `json:"balance"`
	Category       string        `json:"category"`
	CategoryID     string        `json:"categoryId"`
	AccountID      string        `json:"accountId"`
	ProviderCode   string        `json:"providerCode"`
	Type           string        `json:"type"`   // CREDIT, DEBIT
	Status         string        `json:"status"` // POSTED, PENDING
	PaymentData    any           `json:"paymentData"`
	Merchant       *MerchantData `json:"merchant"`
}

type TransactionsResponse struct {
	Total      int           `json:"total"`
	TotalPages int           `json:"totalPages"`
	Page       int           `json:"page"`
	Results    []Transaction `json:"results"`
}

// CreditCardBill representa a fatura de um cartão de crédito
type CreditCardBill struct {
	ID                    string    `json:"id"`
	DueDate               time.Time `json:"dueDate"`
	BillClosingDate       *time.Time `json:"billClosingDate"`
	TotalAmount           float64   `json:"totalAmount"`
	TotalAmountCurrencyCode string  `json:"totalAmountCurrencyCode"`
	MinimumPaymentAmount  *float64  `json:"minimumPaymentAmount"`
	AllowsInstallments    bool      `json:"allowsInstallments"`
	FinanceCharges        any       `json:"financeCharges"`
	Payments              any       `json:"payments"`
}

type BillsResponse struct {
	Total      int              `json:"total"`
	TotalPages int              `json:"totalPages"`
	Page       int              `json:"page"`
	Results    []CreditCardBill `json:"results"`
}

// Investment representa uma aplicação financeira
type Investment struct {
	ID              string     `json:"id"`
	Code            string     `json:"code"`
	ISIN            string     `json:"isin"`
	Name            string     `json:"name"`
	Type            string     `json:"type"` // MUTUAL_FUND, FIXED_INCOME, EQUITY, SECURITY, ETF
	Subtype         string     `json:"subtype"`
	Balance         float64    `json:"balance"`
	Amount          float64    `json:"amount"`
	CurrencyCode    string     `json:"currencyCode"`
	DueDate         *time.Time `json:"dueDate"`
	Rate            *float64   `json:"rate"`
	RateType        string     `json:"rateType"`
	FixedAnnualRate *float64   `json:"fixedAnnualRate"`
	LastMonthRate   *float64   `json:"lastMonthRate"`
	AnnualRate      *float64   `json:"annualRate"`
	Status          string     `json:"status"`
	Institution     string     `json:"institution"`
}

type InvestmentsResponse struct {
	Total      int          `json:"total"`
	TotalPages int          `json:"totalPages"`
	Page       int          `json:"page"`
	Results    []Investment `json:"results"`
}

// InvestmentTransaction representa uma movimentação em um investimento
type InvestmentTransaction struct {
	ID              string     `json:"id"`
	TradeDate       time.Time  `json:"tradeDate"`
	Date            time.Time  `json:"date"`
	Quantity        *float64   `json:"quantity"`
	Value           *float64   `json:"value"`
	Amount          float64    `json:"amount"`
	Type            string     `json:"type"` // BUY, SELL, TAX, INTEREST, AMORTIZATION
	MovementType    string     `json:"movementType"` // CREDIT, DEBIT
	Description     string     `json:"description"`
	BrokerageNumber string     `json:"brokerageNumber"`
	NetAmount       *float64   `json:"netAmount"`
	AgreedRate      *float64   `json:"agreedRate"`
	Expenses        any        `json:"expenses"`
}

type InvestmentTransactionsResponse struct {
	Total      int                     `json:"total"`
	TotalPages int                     `json:"totalPages"`
	Page       int                     `json:"page"`
	Results    []InvestmentTransaction `json:"results"`
}

// WebhookRegistration representa a configuração de um webhook
type WebhookRegistration struct {
	ID      string            `json:"id,omitempty"`
	Event   string            `json:"event"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type WebhooksListResponse struct {
	Total      int                   `json:"total"`
	TotalPages int                   `json:"totalPages"`
	Page       int                   `json:"page"`
	Results    []WebhookRegistration `json:"results"`
}

// WebhookPayload representa o payload recebido quando um webhook dispara
type WebhookPayload struct {
	Event       string `json:"event"`
	EventID     string `json:"eventId"`
	ItemID      string `json:"itemId"`
	AccountID   string `json:"accountId"`
	TriggeredBy string `json:"triggeredBy"`
	Data        any    `json:"data"`
}

