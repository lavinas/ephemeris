package driven

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"billing/internal/domain"
)

const batchSizeInsertTransaction = 100

// PostgresRepository is an adapter for GORM database operations
type PostgresRepository struct {
	DB *gorm.DB
	Tx *gorm.DB
}

// NewPostgresRepository creates a new instance of PostgresRepository
func NewPostgresRepository(host, user, password, dbname, sslmode, timezone string, port,
	timeout int, schema string) (*PostgresRepository, error) {
	rep := &PostgresRepository{DB: nil}
	dns := "postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s&search_path=%s&connect_timeout=%d"
	dns = fmt.Sprintf(dns, user, password, host, port, dbname, sslmode, timezone, schema, timeout)
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
	return db.Ping()
}

// Close closes the database connection
func (a *PostgresRepository) Close() error {
	db, err := a.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// BeginTransaction starts a new database transaction
func (a *PostgresRepository) BeginTransaction() error {
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
func (a *PostgresRepository) CommitTransaction() error {
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
func (a *PostgresRepository) RollbackTransaction() error {
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
func (a *PostgresRepository) Save(model interface{}) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(model, batchSizeInsertTransaction).Error
}

// FindCustomers retrieves customers based on the provided filters and pagination parameters
func (a *PostgresRepository) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	var customers []domain.Customer
	db := a.DB.Model(&domain.Customer{})
	db = db.Where("vendor_id = ?", vendorID)
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
func (a *PostgresRepository) GetCustomer(vendorID int64, nickname string) (*domain.Customer,
	error) {
	var customer domain.Customer
	err := a.DB.Where("vendor_id = ? AND nickname = ?", vendorID, nickname).First(&customer).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // Return nil if no record is found
	}
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindVendors retrieves vendors based on the provided filters and pagination parameters
func (a *PostgresRepository) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	var vendors []domain.Vendor
	db := a.DB.Model(&domain.Vendor{})
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
func (a *PostgresRepository) GetVendor(nickname string) (*domain.Vendor, error) {
	var vendor domain.Vendor
	err := a.DB.Where("nickname = ?", nickname).First(&vendor).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // Return nil if no record is found
	}
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// FindInvoices retrieves invoices based on the provided filters and pagination parameters
func (a *PostgresRepository) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, taxDate *string) ([]domain.Invoice, error) {
	var invoices []domain.Invoice
	db := a.DB.Model(&domain.Invoice{}).Preload("InvoiceItems").
		Preload("Customer")
	if customer != 0 {
		db = db.Where("customer_id = ?", customer)
	}
	if invoiceDate != nil {
		db = db.Where("invoice_date::text ILIKE ?", "%"+*invoiceDate+"%")
	}
	if dueDate != nil {
		db = db.Where("due_date::text ILIKE ?", "%"+*dueDate+"%")
	}
	if paymentDate != nil {
		db = db.Where("payment_date::text ILIKE ?", "%"+*paymentDate+"%")
	}
	if emailSentDate != nil {
		db = db.Where("email_sent_date::text ILIKE ?", "%"+*emailSentDate+"%")
	}
	if whatsappSentDate != nil {
		db = db.Where("whatsapp_sent_date::text ILIKE ?", "%"+*whatsappSentDate+"%")
	}
	if taxDate != nil {
		db = db.Where("tax_date::text ILIKE ?", "%"+*taxDate+"%")
	}
	if page > 0 && pageSize > 0 {
		db = db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err := db.Find(&invoices).Error
	return invoices, err
}

// GetInvoice retrieves a single invoice by ID
func (a *PostgresRepository) GetInvoice(id int64) (*domain.Invoice, error) {
	var invoice domain.Invoice
	err := a.DB.Preload("InvoiceItems").Preload("Customer").
		Where("id = ?", id).First(&invoice).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // Return nil if no record is found
	}
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetInvoicesByKey retrieves invoices based on the customer ID and invoice date
func (a *PostgresRepository) GetInvoicesByKey(customerID int64, invoiceDate time.Time) ([]domain.Invoice, error) {
	var invoices []domain.Invoice
	err := a.DB.Preload("InvoiceItems").Preload("Customer").
		Where("customer_id = ? AND invoice_date::text = ?", customerID, invoiceDate.Format("2006-01-02")).
		Find(&invoices).Error
	return invoices, err
}