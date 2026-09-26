package repository

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

// CountFind is a helper function to count records in the database based on conditions
func (a *Repository) FindCount(conditions map[string]interface{}) (int64, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	for key, value := range conditions {
		if value == nil {
			db = db.Where(key)
		} else {
			db = db.Where(key, value)
		}
	}
	var count int64
	err := db.Find(&domain.Session{}).Count(&count).Error
	return count, err
}

// Find is a helper function to find records in the database
func (a *Repository) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	if len(orderBy) > 0 {
		for _, order := range orderBy {
			if order != "" {
				db = db.Order(order)
			}
		}
	}
	if len(orderBy) > 0 && orderBy[0] != "" {
		db = db.Order(orderBy[0])
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

// FindSessionUsers is a helper function to find session users in the database based on conditions
func (a *Repository) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	md := db.Model(&domain.Session{})
	for key, value := range conditions {
		if value == nil {
			md = md.Where(key)
		} else {
			md = md.Where(key, value)
		}
	}
	qselect := groupField + ", COUNT(1) as count"
	md = md.Select(qselect)
	md = md.Group(groupField)
	var results []map[string]interface{}
	err := md.Find(&results).Error
	return results, err
}

// FindCustomers retrieves customers based on the provided filters and pagination parameters
func (a *Repository) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	var customers []domain.Customer
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	db = db.Model(&domain.Customer{}).Where("vendor_id = ?", vendorID)
	if name != nil {
		db = db.Where("name ILIKE ?", "%"+*name+"%")
	}
	if nickname != nil {
		db = db.Where("nickname ILIKE ?", "%"+*nickname+"%")
	}
	if document != nil {
		db = db.Where("document ILIKE ?", "%"+*document+"%")
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if email != nil {
		db = db.Where("email ILIKE ?", "%"+*email+"%")
	}
	if whatsapp != nil {
		db = db.Where("whatsapp ILIKE ?", "%"+*whatsapp+"%")
	}
	if page > 0 && pageSize > 0 {
		db = db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err := db.Find(&customers).Error
	return customers, err
}

// GetCustomer retrieves a single customer by Nickname
func (a *Repository) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	var customer domain.Customer
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	err := db.Where("vendor_id = ? AND nickname = ?", vendorID, nickname).First(&customer).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindVendors retrieves vendors based on the provided filters and pagination parameters
func (a *Repository) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	var vendors []domain.Vendor
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	db = db.Model(&domain.Vendor{})
	if legalName != nil {
		db = db.Where("legal_name ILIKE ?", "%"+*legalName+"%")
	}
	if nickname != nil {
		db = db.Where("nickname ILIKE ?", "%"+*nickname+"%")
	}
	if document != nil {
		db = db.Where("document ILIKE ?", "%"+*document+"%")
	}
	if accountBank != nil {
		db = db.Where("account_bank ILIKE ?", "%"+*accountBank+"%")
	}
	if accountAgency != nil {
		db = db.Where("account_agency ILIKE ?", "%"+*accountAgency+"%")
	}
	if accountNumber != nil {
		db = db.Where("account_number ILIKE ?", "%"+*accountNumber+"%")
	}
	if page > 0 && pageSize > 0 {
		db = db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err := db.Find(&vendors).Error
	return vendors, err
}

// GetVendor retrieves a single vendor by Nickname
func (a *Repository) GetVendor(nickname string) (*domain.Vendor, error) {
	var vendor domain.Vendor
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	err := db.Where("nickname = ?", nickname).First(&vendor).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}
