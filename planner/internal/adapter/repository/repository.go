package repository

import (
	"fmt"
	"time"

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
	return db.Omit(clause.Associations).Clauses(clause.OnConflict{
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
	err := db.Preload("Customer").Preload("Service").Find(&sessions).Error
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

// GetCustomerByID retrieves a single customer by ID
func (a *Repository) GetCustomerByID(id int64) (*domain.Customer, error) {
	var customer domain.Customer
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	err := db.Where("id = ?", id).First(&customer).Error
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

// GetVendorByID retrieves a single vendor by ID
func (a *Repository) GetVendorByID(id int64) (*domain.Vendor, error) {
	var vendor domain.Vendor
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	err := db.Where("id = ?", id).First(&vendor).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// FindServices retrieves all services for a given vendorID ordered by id
func (a *Repository) FindServices(vendorID int64) ([]domain.Service, error) {
	var services []domain.Service
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	err := db.Where("vendor_id = ?", vendorID).Order("id asc").Find(&services).Error
	return services, err
}

// GetServiceByID retrieves a single service by ID
func (a *Repository) GetServiceByID(id int64) (*domain.Service, error) {
	var service domain.Service
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	err := db.Where("id = ?", id).First(&service).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// SavePlan saves a plan model and its associated items and specialization table
func (a *Repository) SavePlan(plan *domain.Plan) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	if plan.ID == 0 {
		plan.CreatedAt = now
		plan.UpdatedAt = now
		tx := db.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		if err := tx.Omit("Items", "Agenda", "Package", "Notebook", "PayPerView", "Customer").Create(plan).Error; err != nil {
			tx.Rollback()
			return err
		}

		for i := range plan.Items {
			plan.Items[i].PlanID = plan.ID
			plan.Items[i].CreatedAt = now
			plan.Items[i].UpdatedAt = now
			if err := tx.Omit("Service").Create(&plan.Items[i]).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		switch plan.PlanType {
		case int(domain.PlanTypeAgenda):
			if plan.Agenda != nil {
				plan.Agenda.ID = plan.ID
				if err := tx.Create(plan.Agenda).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		case int(domain.PlanTypePackage):
			if plan.Package != nil {
				plan.Package.ID = plan.ID
				if err := tx.Create(plan.Package).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		case int(domain.PlanTypeNotebook):
			if plan.Notebook != nil {
				plan.Notebook.ID = plan.ID
				if err := tx.Create(plan.Notebook).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		case int(domain.PlanTypePayPerView):
			if plan.PayPerView != nil {
				plan.PayPerView.ID = plan.ID
				if err := tx.Create(plan.PayPerView).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		return tx.Commit().Error
	}

	plan.UpdatedAt = now
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	updates := map[string]interface{}{
		"customer_id": plan.CustomerID,
		"plan_start":  plan.PlanStart,
		"plan_end":    plan.PlanEnd,
		"plan_type":   plan.PlanType,
		"price":       plan.Price,
		"updated_at":  now,
	}
	if err := tx.Model(&domain.Plan{}).Where("id = ?", plan.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(plan.Items) > 0 {
		if err := tx.Where("plan_id = ?", plan.ID).Delete(&domain.PlanItem{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		for i := range plan.Items {
			plan.Items[i].PlanID = plan.ID
			plan.Items[i].CreatedAt = now
			plan.Items[i].UpdatedAt = now
			if err := tx.Omit("Service").Create(&plan.Items[i]).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	} else if plan.PlanType == int(domain.PlanTypeManual) {
		if err := tx.Where("plan_id = ?", plan.ID).Delete(&domain.PlanItem{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if plan.PlanType != int(domain.PlanTypeAgenda) {
		_ = tx.Where("id = ?", plan.ID).Delete(&domain.PlanAgenda{}).Error
	}
	if plan.PlanType != int(domain.PlanTypePackage) {
		_ = tx.Where("id = ?", plan.ID).Delete(&domain.PlanPackage{}).Error
	}
	if plan.PlanType != int(domain.PlanTypeNotebook) {
		_ = tx.Where("id = ?", plan.ID).Delete(&domain.PlanNotebook{}).Error
	}
	if plan.PlanType != int(domain.PlanTypePayPerView) {
		_ = tx.Where("id = ?", plan.ID).Delete(&domain.PlanPayPerView{}).Error
	}

	switch plan.PlanType {
	case int(domain.PlanTypeAgenda):
		if plan.Agenda != nil {
			plan.Agenda.ID = plan.ID
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(plan.Agenda).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	case int(domain.PlanTypePackage):
		if plan.Package != nil {
			plan.Package.ID = plan.ID
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(plan.Package).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	case int(domain.PlanTypeNotebook):
		if plan.Notebook != nil {
			plan.Notebook.ID = plan.ID
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(plan.Notebook).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	case int(domain.PlanTypePayPerView):
		if plan.PayPerView != nil {
			plan.PayPerView.ID = plan.ID
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(plan.PayPerView).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// GetPlanByID retrieves a single plan by ID with preloaded associations
func (a *Repository) GetPlanByID(id int64) (*domain.Plan, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var plan domain.Plan
	err := db.Preload("Customer").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("order_index asc")
		}).
		Preload("Items.Service").
		Preload("Agenda").
		Preload("Package").
		Preload("Notebook").
		Preload("PayPerView").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&plan).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// DeletePlan performs soft delete on a plan and its items
func (a *Repository) DeletePlan(id int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Model(&domain.Plan{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&domain.PlanItem{}).Where("plan_id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// FindPlans retrieves plans based on conditions, pagination and ordering
func (a *Repository) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	query := db.Model(&domain.Plan{})
	for k, v := range conditions {
		if v == nil {
			query = query.Where(k)
		} else {
			query = query.Where(k, v)
		}
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if len(orderBy) > 0 {
		for _, o := range orderBy {
			if o != "" {
				query = query.Order(o)
			}
		}
	} else {
		query = query.Order("plan_start desc").Order("id desc")
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	var plans []domain.Plan
	err := query.Preload("Customer").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("order_index asc")
		}).
		Preload("Items.Service").
		Preload("Agenda").
		Preload("Package").
		Preload("Notebook").
		Preload("PayPerView").
		Find(&plans).Error
	return plans, count, err
}

// FindCustomerActivePlansWithServices finds all active non-deleted plans for a customer that include specified services
func (a *Repository) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var plans []domain.Plan
	query := db.Model(&domain.Plan{}).
		Joins("JOIN planner.plan_item ON planner.plan_item.plan_id = planner.plan.id").
		Where("planner.plan.customer_id = ? AND planner.plan.deleted_at IS NULL AND planner.plan_item.deleted_at IS NULL", customerID).
		Where("planner.plan_item.service_id IN (?)", serviceIDs)
	if excludePlanID > 0 {
		query = query.Where("planner.plan.id != ?", excludePlanID)
	}
	query = query.Distinct().
		Preload("Customer").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("order_index asc")
		}).
		Preload("Items.Service")
	err := query.Find(&plans).Error
	return plans, err
}

// SaveInvoice saves an invoice, its items and session links in a transaction
func (a *Repository) SaveInvoice(invoice *domain.Invoice) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	invoice.UpdatedAt = now
	if invoice.CreatedAt.IsZero() {
		invoice.CreatedAt = now
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Omit("Items", "Plan").Create(invoice).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range invoice.Items {
		invoice.Items[i].InvoiceID = invoice.ID
		invoice.Items[i].CreatedAt = now
		invoice.Items[i].UpdatedAt = now
		if invoice.Items[i].Quantity == 0 && invoice.Items[i].InvoicedQuantity > 0 {
			invoice.Items[i].Quantity = invoice.Items[i].InvoicedQuantity
		}
		if err := tx.Omit("Service", "Invoice").Create(&invoice.Items[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetInvoiceByID retrieves an invoice by ID
func (a *Repository) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var invoice domain.Invoice
	err := db.Preload("Plan").
		Preload("Plan.Customer").
		Preload("Plan.Customer.Vendor").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL")
		}).
		Preload("Items.Service").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&invoice).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindInvoices retrieves invoices matching conditions
func (a *Repository) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	query := db.Model(&domain.Invoice{})
	for k, v := range conditions {
		if v == nil {
			query = query.Where(k)
		} else {
			query = query.Where(k, v)
		}
	}
	if len(orderBy) > 0 {
		for _, o := range orderBy {
			if o != "" {
				query = query.Order(o)
			}
		}
	} else {
		query = query.Order("invoicing_date desc, id desc")
	}

	var invoices []domain.Invoice
	err := query.Preload("Plan").
		Preload("Plan.Customer").
		Preload("Plan.Customer.Vendor").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL")
		}).
		Preload("Items.Service").
		Find(&invoices).Error
	return invoices, err
}

// FindInvoicesPaged retrieves invoices matching conditions with pagination, ordering, total count and total amount
func (a *Repository) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	query := db.Model(&domain.Invoice{})
	for k, v := range conditions {
		if v == nil {
			query = query.Where(k)
		} else {
			query = query.Where(k, v)
		}
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, 0, err
	}

	var totalAmount float64
	if err := query.Session(&gorm.Session{}).Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalAmount); err != nil {
		return nil, 0, 0, err
	}

	if len(orderBy) > 0 {
		for _, o := range orderBy {
			if o != "" {
				query = query.Order(o)
			}
		}
	} else {
		query = query.Order("invoicing_date desc").Order("id desc")
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	var invoices []domain.Invoice
	err := query.Preload("Plan").
		Preload("Plan.Customer").
		Preload("Plan.Customer.Vendor").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("id asc")
		}).
		Preload("Items.Service").
		Find(&invoices).Error
	return invoices, count, totalAmount, err
}

// DeleteInvoice soft-deletes an invoice and its items
func (a *Repository) DeleteInvoice(id int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Model(&domain.Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&domain.InvoiceItem{}).Where("invoice_id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// UpdateInvoice updates an invoice's due date, amount, and items' invoiced quantities in a transaction.
func (a *Repository) UpdateInvoice(invoice *domain.Invoice) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	invoice.UpdatedAt = now
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	invoiceUpdates := map[string]interface{}{
		"due_date":   invoice.DueDate,
		"amount":     invoice.Amount,
		"updated_at": now,
	}
	if invoice.ReferenceInvoiceID != nil {
		invoiceUpdates["reference_invoice_id"] = invoice.ReferenceInvoiceID
	}
	if err := tx.Model(&domain.Invoice{}).Where("id = ? AND deleted_at IS NULL", invoice.ID).Updates(invoiceUpdates).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, it := range invoice.Items {
		if it.ID > 0 {
			qty := it.Quantity
			if it.InvoicedQuantity > 0 && (it.Quantity == 0 || it.InvoicedQuantity != it.Quantity) {
				qty = it.InvoicedQuantity
			}
			if err := tx.Model(&domain.InvoiceItem{}).Where("id = ? AND invoice_id = ? AND deleted_at IS NULL", it.ID, invoice.ID).Updates(map[string]interface{}{
				"quantity":   qty,
				"updated_at": now,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// DeleteInvoicesByIDs soft-deletes multiple invoices and their items in a single transaction.
func (a *Repository) DeleteInvoicesByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&domain.Invoice{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&domain.InvoiceItem{}).Where("invoice_id IN ?", ids).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// FindActivePlansForInvoicing finds all non-deleted plans active in the given period matching filters.
// For post-paid plans (agenda post-paid, notebook), sessions from the previous month are billed in the current month,
// so plans active in the month prior to startDate are included.
func (a *Repository) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	firstOfStartMonth := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	minActiveDate := firstOfStartMonth.AddDate(0, -1, 0)

	query := db.Model(&domain.Plan{}).
		Where("planner.plan.deleted_at IS NULL").
		Where("planner.plan.plan_start <= ?", endDate).
		Where("(planner.plan.plan_end IS NULL OR planner.plan.plan_end >= ?)", minActiveDate)

	if planID != nil && *planID != 0 {
		query = query.Where("planner.plan.id = ?", *planID)
	}
	if vendorID != nil && *vendorID != 0 {
		query = query.Where("planner.plan.customer_id IN (SELECT id FROM planner.customer WHERE vendor_id = ? AND deleted_at IS NULL)", *vendorID)
	}
	if customerID != nil && *customerID != 0 {
		query = query.Where("planner.plan.customer_id = ?", *customerID)
	}
	if planType != nil && *planType != 0 {
		query = query.Where("planner.plan.plan_type = ?", *planType)
	}
	if serviceID != nil && *serviceID != 0 {
		query = query.Where("planner.plan.id IN (SELECT plan_id FROM planner.plan_item WHERE service_id = ? AND deleted_at IS NULL)", *serviceID)
	}

	var plans []domain.Plan
	err := query.Preload("Customer").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("order_index asc")
		}).
		Preload("Items.Service").
		Preload("Agenda").
		Preload("Package").
		Preload("Notebook").
		Preload("PayPerView").
		Order("planner.plan.plan_start asc, planner.plan.id asc").
		Find(&plans).Error
	return plans, err
}

// FindBillableSessions finds non-deleted, billable sessions for a customer and service in a date range
func (a *Repository) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var sessions []domain.Session
	query := db.Model(&domain.Session{}).
		Where("customer_id = ? AND service_id = ?", customerID, serviceID).
		Where("session_date >= ? AND session_date <= ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Where("session_status NOT IN (?)", []string{string(domain.StatusCancelNotCharge), "Cancelada/Não Cobrar"}).
		Order("session_date asc, id asc")

	err := query.Preload("Customer").Preload("Service").Find(&sessions).Error
	return sessions, err
}

// GetInvoiceForSession finds active invoice for a session
func (a *Repository) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var invoice domain.Invoice
	err := db.Model(&domain.Invoice{}).
		Joins("JOIN planner.invoice_item ON planner.invoice_item.invoice_id = planner.invoice.id").
		Joins("JOIN planner.invoice_item_session ON planner.invoice_item_session.invoice_item_id = planner.invoice_item.id").
		Where("planner.invoice_item_session.session_id = ? AND planner.invoice.deleted_at IS NULL AND planner.invoice_item.deleted_at IS NULL", sessionID).
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL")
		}).
		First(&invoice).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// SaveInvoices saves multiple invoices and their items in a transaction
func (a *Repository) SaveInvoices(invoices []*domain.Invoice) error {
	db := a.DB
	useExistingTx := false
	if a.Tx != nil {
		db = a.Tx
		useExistingTx = true
	}

	tx := db
	if !useExistingTx {
		tx = db.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
	}

	now := time.Now()
	for _, invoice := range invoices {
		invoice.UpdatedAt = now
		if invoice.CreatedAt.IsZero() {
			invoice.CreatedAt = now
		}

		if err := tx.Omit("Items", "Plan").Create(invoice).Error; err != nil {
			if !useExistingTx {
				tx.Rollback()
			}
			return err
		}

		for i := range invoice.Items {
			invoice.Items[i].InvoiceID = invoice.ID
			invoice.Items[i].CreatedAt = now
			invoice.Items[i].UpdatedAt = now
			if invoice.Items[i].Quantity == 0 && invoice.Items[i].InvoicedQuantity > 0 {
				invoice.Items[i].Quantity = invoice.Items[i].InvoicedQuantity
			}
			if err := tx.Omit("Service", "Invoice").Create(&invoice.Items[i]).Error; err != nil {
				if !useExistingTx {
					tx.Rollback()
				}
				return err
			}
		}
	}

	if !useExistingTx {
		return tx.Commit().Error
	}
	return nil
}

// CheckInvoiceExists checks if an invoice exists for the same customer, invoice date, period start and period end
func (a *Repository) CheckInvoiceExists(customerID int64, invoiceDate, periodStart, periodEnd time.Time, excludeInvoiceID int64) (bool, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var count int64
	query := db.Table("planner.invoice").
		Joins("JOIN planner.plan ON planner.plan.id = planner.invoice.plan_id").
		Where("planner.plan.customer_id = ?", customerID).
		Where("planner.invoice.invoicing_date = ?", invoiceDate.Format("2006-01-02")).
		Where("planner.invoice.period_start = ?", periodStart.Format("2006-01-02")).
		Where("planner.invoice.period_end = ?", periodEnd.Format("2006-01-02")).
		Where("planner.invoice.deleted_at IS NULL").
		Where("planner.plan.deleted_at IS NULL")

	if excludeInvoiceID > 0 {
		query = query.Where("planner.invoice.id <> ?", excludeInvoiceID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}


