package driven

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const batchSizeInsertTransaction = 100

// PostgresRepository is an adapter for GORM database operations
type PostgresRepository struct {
	DB  *gorm.DB
	ctx *context.Context
}

// NewPostgresRepository creates a new instance of PostgresRepository
func NewPostgresRepository(host, user, password, dbname, sslmode, timezone string, port, connect_timeout int,
	billing_schema string, ctx *context.Context) (*PostgresRepository, error) {
	rep := &PostgresRepository{DB: nil, ctx: ctx}
	dns := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s search_path=%s connect_timeout=%d",
		host, port, user, password, dbname, sslmode, timezone, billing_schema, connect_timeout)
	if err := rep.Connect(dns); err != nil {
		return nil, err
	}
	return rep, nil
}

// Connect establishes a connection to the database (placeholder for actual connection logic)
func (a *PostgresRepository) Connect(dns string) error {
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
func (a *PostgresRepository) Ping() error {
	db, err := a.DB.DB()
	if err != nil {
		return err
	}
	return db.PingContext(*a.ctx)
}

// Close closes the database connection
func (a *PostgresRepository) Close() error {
	db, err := a.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// GeneralSave is a helper function to save records to the database with conflict handling
func (a *PostgresRepository) Save(model interface{}) error {
	return a.DB.WithContext(*a.ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(model, batchSizeInsertTransaction).Error
}
