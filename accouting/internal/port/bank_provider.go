package port

import (
	"context"
	"time"

	"accounting/internal/domain"
)

// BankTransaction represents a transaction returned by Santander Open API Contas Correntes
type BankTransaction struct {
	CreditDebitType    string    // "DEBITO" or "CREDITO"
	TransactionName    string    // Description / label
	HistoricComplement string    // Complement / document
	Amount             float64   // Raw positive amount
	TransactionDate    time.Time // Date of the transaction
	DocumentNumber     string    // Document number if provided
}

// BankStatementProvider defines the contract for bank statement providers (Santander, OFX, etc.)
type BankStatementProvider interface {
	// ProviderName returns the identifier of the provider (e.g. "santander_api", "ofx")
	ProviderName() string
	// FetchStatements retrieves transactions for a given account within a date window
	FetchStatements(ctx context.Context, account *domain.Account, creds *domain.AccountSantanderAPI, initialDate, finalDate time.Time) ([]BankTransaction, error)
}
