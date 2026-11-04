package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"accounting/internal/domain"
	"accounting/internal/port"
)

// Repository implements port.Repository using GORM
type Repository struct {
	db *gorm.DB
}

var _ port.Repository = (*Repository)(nil)

// NewRepository connects to postgres and returns a repository
func NewRepository(host, user, password, dbname, sslmode, timezone string, portNum, timeout int, schema string) (*Repository, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s&search_path=%s&connect_timeout=%d",
		user, password, host, portNum, dbname, sslmode, timezone, schema, timeout)

	cfg := &gorm.Config{
		Logger:      gormlogger.Default.LogMode(gormlogger.Silent),
		PrepareStmt: true,
	}

	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() {
	if r.db != nil {
		sqlDB, err := r.db.DB()
		if err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}

// GetAccounts returns accounts with optional filter for active-only
func (r *Repository) GetAccounts(activeOnly bool) ([]domain.Account, error) {
	var accounts []domain.Account
	q := r.db.Preload("AccountSantanderAPI").Preload("Vendor").Where("deleted_at IS NULL")
	if activeOnly {
		q = q.Where("status = ?", 1)
	}
	err := q.Order("id asc").Find(&accounts).Error
	return accounts, err
}

// GetAccountByID returns an account with its relationships
func (r *Repository) GetAccountByID(id int64) (*domain.Account, error) {
	var account domain.Account
	err := r.db.Preload("AccountSantanderAPI").Preload("Vendor").
		Where("id = ? AND deleted_at IS NULL", id).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// GetAccountSantanderAPI returns Santander API credentials for an account
func (r *Repository) GetAccountSantanderAPI(accountID int64) (*domain.AccountSantanderAPI, error) {
	var creds domain.AccountSantanderAPI
	err := r.db.Where("id = ?", accountID).First(&creds).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &creds, nil
}

// GetVendors returns all vendors
func (r *Repository) GetVendors() ([]domain.Vendor, error) {
	var vendors []domain.Vendor
	err := r.db.Order("id asc").Find(&vendors).Error
	return vendors, err
}

// ListStatements returns statements for an account ordered by sdate desc, id desc with pagination
func (r *Repository) ListStatements(accountID int64, startDate, endDate *time.Time, offset, limit int) ([]domain.Statement, int64, error) {
	var statements []domain.Statement
	var total int64

	q := r.db.Model(&domain.Statement{}).Where("account_id = ? AND deleted_at IS NULL", accountID)
	if startDate != nil {
		q = q.Where("sdate >= ?", startDate)
	}
	if endDate != nil {
		q = q.Where("sdate <= ?", endDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Order("sdate desc, id desc").Offset(offset).Limit(limit).Find(&statements).Error
	if err != nil {
		return nil, 0, err
	}

	return statements, total, nil
}

// GetStatementByChecksum checks whether a statement with the given checksum already exists
func (r *Repository) GetStatementByChecksum(checksum string) (*domain.Statement, error) {
	var st domain.Statement
	err := r.db.Where("checksum = ? AND deleted_at IS NULL", checksum).First(&st).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &st, nil
}

// GetSantanderStatementByTransactionID checks whether a Santander transaction was already imported
func (r *Repository) GetSantanderStatementByTransactionID(transactionID string) (*domain.StatementSantanderAPI, error) {
	var s domain.StatementSantanderAPI
	err := r.db.Where("transaction_id = ? AND deleted_at IS NULL", transactionID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// CreateSantanderStatement atomically persists statement and statement_santander_api in a transaction
func (r *Repository) CreateSantanderStatement(st *domain.Statement, santanderSt *domain.StatementSantanderAPI) error {
	now := time.Now()
	st.CreatedAt = now
	st.UpdatedAt = now

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Insert statement
		if err := tx.Create(st).Error; err != nil {
			return err
		}

		// 2. Set foreign key ID in statement_santander_api
		santanderSt.ID = st.ID
		santanderSt.CreatedAt = now
		santanderSt.UpdatedAt = now

		if err := tx.Create(santanderSt).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateStatement inserts a standard statement entry
func (r *Repository) CreateStatement(st *domain.Statement) error {
	now := time.Now()
	st.CreatedAt = now
	st.UpdatedAt = now
	return r.db.Create(st).Error
}

