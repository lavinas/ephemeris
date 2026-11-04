package port

import (
	"context"
	"time"

	"accounting/internal/domain"
)

// BankTransaction represents a normalized transaction returned by any bank provider
type BankTransaction struct {
	TransactionID      string
	TransactionStatus  string
	TransactionSign    string  // "+" or "-"
	TransactionName    string  // Description or label
	TransactionType    string
	TransactionAmount  float64 // Absolute amount
	TransactionCurrency string
	TransactionDate    time.Time
	PartieDocument     string
	PartieType         string
	PartieBankCode     string
	PartieBranchCode   string
	PartieAccount      string
}

// BankStatementProvider defines the contract for bank statement providers (Santander, OFX, etc.)
type BankStatementProvider interface {
	// ProviderName returns the identifier of the provider (e.g. "santander_api", "ofx")
	ProviderName() string
	// FetchStatements retrieves transactions for a given account within a date window
	FetchStatements(ctx context.Context, account *domain.Account, creds *domain.AccountSantanderAPI, initialDate, finalDate time.Time) ([]BankTransaction, error)
}

