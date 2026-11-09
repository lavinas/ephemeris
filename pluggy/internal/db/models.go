package db

import (
	"time"
)

// DBItem representa o registro na tabela pluggy_item
type DBItem struct {
	ID                   string     `json:"id"`
	ConnectorID          int64      `json:"connector_id"`
	ConnectorName        string     `json:"connector_name"`
	ConnectorImageURL    string     `json:"connector_image_url"`
	ConnectorPrimaryColor string    `json:"connector_primary_color"`
	Status               string     `json:"status"`
	ExecutionStatus      string     `json:"execution_status"`
	ClientUserID         string     `json:"client_user_id"`
	ErrorCode            string     `json:"error_code"`
	ErrorMessage         string     `json:"error_message"`
	LastUpdatedAt        *time.Time `json:"last_updated_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Relacionamentos para listagem no painel
	AccountsCount     int `json:"accounts_count"`
	InvestmentsCount  int `json:"investments_count"`
	TransactionsCount int `json:"transactions_count"`
}

// DBAccount representa o registro na tabela pluggy_account
type DBAccount struct {
	ID                             string     `json:"id"`
	ItemID                         string     `json:"item_id"`
	Type                           string     `json:"type"`    // BANK, CREDIT
	Subtype                        string     `json:"subtype"` // CHECKING_ACCOUNT, SAVINGS_ACCOUNT, CREDIT_CARD
	Name                           string     `json:"name"`
	MarketingName                  string     `json:"marketing_name"`
	Balance                        float64    `json:"balance"`
	CurrencyCode                   string     `json:"currency_code"`
	Number                         string     `json:"number"`
	Owner                          string     `json:"owner"`
	TaxNumber                      string     `json:"tax_number"`
	BankAgency                     string     `json:"bank_agency"`
	BankTransferNumber             string     `json:"bank_transfer_number"`
	BankClosingBalance             *float64   `json:"bank_closing_balance"`
	BankOverdraftContractedLimit   *float64   `json:"bank_overdraft_contracted_limit"`
	BankOverdraftUsedLimit         *float64   `json:"bank_overdraft_used_limit"`
	BankAutomaticallyInvestedBalance *float64 `json:"bank_automatically_invested_balance"`
	BankHasReservedBalance         bool       `json:"bank_has_reserved_balance"`
	CreditBrand                    string     `json:"credit_brand"`
	CreditLevel                    string     `json:"credit_level"`
	CreditLimit                    *float64   `json:"credit_limit"`
	CreditAvailableLimit           *float64   `json:"credit_available_limit"`
	CreditBalanceCloseDate         *time.Time `json:"credit_balance_close_date"`
	CreditBalanceDueDate           *time.Time `json:"credit_balance_due_date"`
	CreditMinimumPayment           *float64   `json:"credit_minimum_payment"`
	CreditStatus                   string     `json:"credit_status"`
	CreditHolderType               string     `json:"credit_holder_type"`
	RawData                        string     `json:"raw_data"`
	CreatedAt                      time.Time  `json:"created_at"`
	UpdatedAt                      time.Time  `json:"updated_at"`

	// Metadados do banco/emissor
	BankName string `json:"bank_name"`
}

// DBTransaction representa o registro na tabela pluggy_transaction
type DBTransaction struct {
	ID             string     `json:"id"`
	AccountID      string     `json:"account_id"`
	ItemID         string     `json:"item_id"`
	Description    string     `json:"description"`
	DescriptionRaw string     `json:"description_raw"`
	Amount         float64    `json:"amount"`
	CurrencyCode   string     `json:"currency_code"`
	Date           time.Time  `json:"date"`
	Category       string     `json:"category"`
	CategoryID     string     `json:"category_id"`
	Type           string     `json:"type"`   // CREDIT, DEBIT
	Status         string     `json:"status"` // POSTED, PENDING
	Balance        *float64   `json:"balance"`
	ProviderCode   string     `json:"provider_code"`
	MerchantName   string     `json:"merchant_name"`
	MerchantCNPJ   string     `json:"merchant_cnpj"`
	PaymentData    string     `json:"payment_data"`
	RawData        string     `json:"raw_data"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Nomes auxiliares
	AccountName    string `json:"account_name"`
	AccountSubtype string `json:"account_subtype"`
	BankName       string `json:"bank_name"`
}

// DBCreditCardBill representa a fatura na tabela pluggy_credit_card_bill
type DBCreditCardBill struct {
	ID                   string     `json:"id"`
	AccountID            string     `json:"account_id"`
	DueDate              time.Time  `json:"due_date"`
	BillClosingDate      *time.Time `json:"bill_closing_date"`
	TotalAmount          float64    `json:"total_amount"`
	CurrencyCode         string     `json:"currency_code"`
	MinimumPaymentAmount *float64   `json:"minimum_payment_amount"`
	AllowsInstallments   bool       `json:"allows_installments"`
	FinanceCharges       string     `json:"finance_charges"`
	Payments             string     `json:"payments"`
	RawData              string     `json:"raw_data"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	CardName string `json:"card_name"`
}

// DBInvestment representa o registro na tabela pluggy_investment
type DBInvestment struct {
	ID              string     `json:"id"`
	ItemID          string     `json:"item_id"`
	Name            string     `json:"name"`
	Code            string     `json:"code"`
	ISIN            string     `json:"isin"`
	Type            string     `json:"type"`
	Subtype         string     `json:"subtype"`
	Balance         float64    `json:"balance"`
	Amount          float64    `json:"amount"`
	CurrencyCode    string     `json:"currency_code"`
	DueDate         *time.Time `json:"due_date"`
	Rate            *float64   `json:"rate"`
	RateType        string     `json:"rate_type"`
	FixedAnnualRate *float64   `json:"fixed_annual_rate"`
	LastMonthRate   *float64   `json:"last_month_rate"`
	AnnualRate      *float64   `json:"annual_rate"`
	Status          string     `json:"status"`
	Institution     string     `json:"institution"`
	RawData         string     `json:"raw_data"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	InstitutionName string `json:"institution_name"`
}

// DBInvestmentTransaction representa a tabela pluggy_investment_transaction
type DBInvestmentTransaction struct {
	ID              string    `json:"id"`
	InvestmentID    string    `json:"investment_id"`
	TradeDate       time.Time `json:"trade_date"`
	Date            time.Time `json:"date"`
	Type            string    `json:"type"`
	MovementType    string    `json:"movement_type"`
	Description     string    `json:"description"`
	Quantity        *float64  `json:"quantity"`
	Value           *float64  `json:"value"`
	Amount          float64   `json:"amount"`
	NetAmount       *float64  `json:"net_amount"`
	AgreedRate      *float64  `json:"agreed_rate"`
	BrokerageNumber string    `json:"brokerage_number"`
	Expenses        string    `json:"expenses"`
	RawData         string    `json:"raw_data"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	InvestmentName string `json:"investment_name"`
}

// DBWebhookEvent representa os eventos gravados em pluggy_webhook_event
type DBWebhookEvent struct {
	ID           int64     `json:"id"`
	Event        string    `json:"event"`
	EventID      string    `json:"event_id"`
	ItemID       string    `json:"item_id"`
	AccountID    string    `json:"account_id"`
	TriggeredBy  string    `json:"triggered_by"`
	Payload      string    `json:"payload"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message"`
	ReceivedAt   time.Time `json:"received_at"`
}

// PortfolioSummary representa o resultado da view vw_portfolio_summary
type PortfolioSummary struct {
	TotalBankBalance         float64 `json:"total_bank_balance"`
	TotalCreditSpent         float64 `json:"total_credit_spent"`
	TotalInvestmentsBalance  float64 `json:"total_investments_balance"`
	TotalConnectedItems      int     `json:"total_connected_items"`
	TotalBankAccounts        int     `json:"total_bank_accounts"`
	TotalCreditCards         int     `json:"total_credit_cards"`
	TotalInvestments         int     `json:"total_investments"`
}

