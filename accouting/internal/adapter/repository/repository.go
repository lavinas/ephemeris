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
func (r *Repository) ListStatements(accountID int64, startDate, endDate *time.Time, txType string, offset, limit int) ([]domain.Statement, int64, error) {
	var statements []domain.Statement
	var total int64

	q := r.db.Model(&domain.Statement{}).Where("account_id = ? AND deleted_at IS NULL", accountID)
	if startDate != nil {
		q = q.Where("sdate >= ?", startDate)
	}
	if endDate != nil {
		q = q.Where("sdate <= ?", endDate)
	}

	if txType == "credit" {
		q = q.Where("amount > 0")
	} else if txType == "debit" {
		q = q.Where("amount < 0")
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

// GetStatementByKey checks whether a statement with the given key already exists
func (r *Repository) GetStatementByKey(key string) (*domain.Statement, error) {
	var st domain.Statement
	err := r.db.Where("key = ? AND deleted_at IS NULL", key).First(&st).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &st, nil
}

// GetSantanderKeyCountForAccountAndDate counts existing Santander statements for the same account, date and santanderKey
func (r *Repository) GetSantanderKeyCountForAccountAndDate(accountID int64, date time.Time, santanderKey string) (int, error) {
	var count int64
	dateStr := date.Format("2006-01-02")
	err := r.db.Table("statement").
		Joins("JOIN statement_santander_api ON statement_santander_api.id = statement.id").
		Where("statement.account_id = ? AND statement.sdate = ? AND statement_santander_api.key = ? AND statement.deleted_at IS NULL", accountID, dateStr, santanderKey).
		Count(&count).Error
	return int(count), err
}

// GetLastStatementBeforeOrOnDate returns the last statement recorded on or before date for accountID (ordered by sdate desc, id desc)
func (r *Repository) GetLastStatementBeforeOrOnDate(accountID int64, date time.Time) (*domain.Statement, error) {
	var st domain.Statement
	dateStr := date.Format("2006-01-02")
	err := r.db.Where("account_id = ? AND sdate <= ? AND deleted_at IS NULL", accountID, dateStr).
		Order("sdate desc, id desc").
		First(&st).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &st, nil
}

// CreateSantanderStatement atomically persists statement and statement_santander_api in a transaction
// and updates the balance of any posterior records.
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

		if err := tx.Create(santanderSt).Error; err != nil {
			return err
		}

		// 3. Update posterior records balance if any
		dateStr := st.SDate.Format("2006-01-02")
		if err := tx.Model(&domain.Statement{}).
			Where("account_id = ? AND deleted_at IS NULL AND (sdate > ? OR (sdate = ? AND id > ?))", st.AccountID, dateStr, dateStr, st.ID).
			Update("balance", gorm.Expr("balance + ?", st.Amount)).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateStatement inserts a standard statement entry and updates posterior records balance
func (r *Repository) CreateStatement(st *domain.Statement) error {
	now := time.Now()
	st.CreatedAt = now
	st.UpdatedAt = now

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(st).Error; err != nil {
			return err
		}

		dateStr := st.SDate.Format("2006-01-02")
		if err := tx.Model(&domain.Statement{}).
			Where("account_id = ? AND deleted_at IS NULL AND (sdate > ? OR (sdate = ? AND id > ?))", st.AccountID, dateStr, dateStr, st.ID).
			Update("balance", gorm.Expr("balance + ?", st.Amount)).Error; err != nil {
			return err
		}

		return nil
	})
}

