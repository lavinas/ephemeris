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
	ListStatements(accountID int64, startDate, endDate *time.Time, txType string, offset, limit int) ([]domain.Statement, int64, error)
	GetStatementByKey(key string) (*domain.Statement, error)

	// Balance and occurrence queries for synchronization
	GetSantanderKeyCountForAccountAndDate(accountID int64, date time.Time, santanderKey string) (int, error)
	GetLastStatementBeforeOrOnDate(accountID int64, date time.Time) (*domain.Statement, error)

	// Atomic persistence of a Santander transaction (statement + statement_santander_api)
	CreateSantanderStatement(st *domain.Statement, santanderSt *domain.StatementSantanderAPI) error

	// General Statement insert (for OFX or other providers)
	CreateStatement(st *domain.Statement) error

	// Statement deletion by date range (for clean re-sync)
	DeleteStatementsByDateRange(accountID int64, startDate, endDate time.Time) (int, error)

	// Action & Action Rules operations
	GetAccountActions(accountID int64) ([]domain.BankAccountAction, error)
	GetActionByID(id int64) (*domain.BankAccountAction, error)
	CreateAccountAction(action *domain.BankAccountAction) error
	UpdateAccountAction(action *domain.BankAccountAction) error
	DeleteAccountAction(id int64) error

	Close()
}
