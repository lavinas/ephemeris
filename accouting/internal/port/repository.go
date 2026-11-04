package port

import (
	"time"

	"accounting/internal/domain"
)

// Repository defines data access operations for banking and statements
type Repository interface {
	// Account & Vendor operations
	GetAccounts(activeOnly bool) ([]domain.Account, error)
	GetAccountByID(id int64) (*domain.Account, error)
	GetAccountSantanderAPI(accountID int64) (*domain.AccountSantanderAPI, error)
	GetVendors() ([]domain.Vendor, error)

	// Statement operations
	ListStatements(accountID int64, startDate, endDate *time.Time, offset, limit int) ([]domain.Statement, int64, error)
	GetStatementByChecksum(checksum string) (*domain.Statement, error)
	GetSantanderStatementByTransactionID(transactionID string) (*domain.StatementSantanderAPI, error)
	
	// Atomic persistence of a Santander transaction (statement + statement_santander_api)
	CreateSantanderStatement(st *domain.Statement, santanderSt *domain.StatementSantanderAPI) error

	// General Statement insert (for OFX or other providers)
	CreateStatement(st *domain.Statement) error

	Close()
}

