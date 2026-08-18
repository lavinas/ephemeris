package adapter

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"planner/internal/domain"
)

const batchSizeInsertTransaction = 100

// Repository is an adapter for GORM database operations
type Repository struct {
	DB *gorm.DB
	Tx *gorm.DB
}

// NewRepository creates a new instance of Repository
func NewRepository(host, user, password, dbname, sslmode, timezone string, port,
	timeout int, schema string) (*Repository, error) {
	rep := &Repository{DB: nil}
	dns := "postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s&search_path=%s&connect_timeout=%d"
	dns = fmt.Sprintf(dns, user, password, host, port, dbname, sslmode, timezone, schema, timeout)
	if err := rep.Connect(dns); err != nil {
		return nil, err
	}
	return rep, nil
}

// Connect establishes a connection to the database (placeholder for actual connection logic)
func (a *Repository) Connect(dns string) error {
	// Placeholder for actual connection logic, using GORM to connect to the database
	gConfig := gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent), // Disables all SQL logging
		PrepareStmt: true,
	}
	sqlDB, err := gorm.Open(postgres.Open(dns), &gConfig)
	if err != nil {
		return err
	}
	a.DB = sqlDB
	// Verify the connection by pinging the database
	return a.Ping()
}

// Ping checks the database connection
func (a *Repository) Ping() error {
	db, err := a.DB.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

// Close closes the database connection
func (a *Repository) Close() error {
	db, err := a.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// BeginTransaction starts a new database transaction
func (a *Repository) BeginTransaction() error {
	if a.Tx != nil {
		return fmt.Errorf("transaction already in progress")
	}
	tx := a.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	a.Tx = tx
	return nil
}

// CommitTransaction commits the given transaction
func (a *Repository) CommitTransaction() error {
	if a.Tx == nil {
		return fmt.Errorf("no transaction in progress")
	}
	if err := a.Tx.Commit().Error; err != nil {
		return err
	}
	a.Tx = nil
	return nil
}

// RollbackTransaction rolls back the given transaction
func (a *Repository) RollbackTransaction() error {
	if a.Tx == nil {
		return fmt.Errorf("no transaction in progress")
	}
	if err := a.Tx.Rollback().Error; err != nil {
		return err
	}
	a.Tx = nil
	return nil
}

// GeneralSave is a helper function to save records to the database with conflict handling
func (a *Repository) Save(model interface{}) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(model, batchSizeInsertTransaction).Error
}

// Find is a helper function to find records in the database
func (a *Repository) Find(page, pagesize int, conditions map[string]interface{}) ([]interface{}, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	if page > 0 && pagesize > 0 {
		offset := (page - 1) * pagesize
		db = db.Offset(offset).Limit(pagesize)
	}
	for key, value := range conditions {
		if value == nil {
			db = db.Where(key)
		} else {
			db = db.Where(key, value)
		}
	}
	var sessions []domain.Session
	err := db.Find(&sessions).Error
	// Convert []domain.Session to []interface{}
	result := make([]interface{}, len(sessions))
	for i, v := range sessions {
		result[i] = v
	}
	return result, err
}
