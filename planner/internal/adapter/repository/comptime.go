package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"planner/internal/domain"
	"planner/internal/port"
)

// LockCustomerService acquires an advisory transaction lock for the customer and service
// when running against PostgreSQL, serializing balance updates for that customer/service.
func (a *Repository) LockCustomerService(ctx context.Context, customerID, serviceID int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	if db.Dialector.Name() == "postgres" {
		return db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(hashtext('comptime_' || ? || '_' || ?))", customerID, serviceID).Error
	}
	return nil
}

// FindOrCreateComptime finds or creates a comptime record for the given vendor, customer, service and month.
// If created, its balance is initialized from the latest previous month's balance (if any).
func (a *Repository) FindOrCreateComptime(ctx context.Context, vendorID, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	normalizedMonth := domain.NormalizeMonth(month)
	monthStr := normalizedMonth.Format("2006-01-02")

	var comptime domain.Comptime
	err := db.WithContext(ctx).
		Where("vendor_id = ? AND customer_id = ? AND service_id = ? AND month = ? AND deleted_at IS NULL",
			vendorID, customerID, serviceID, monthStr).
		First(&comptime).Error

	if err == nil {
		return &comptime, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Calculate initial balance based on previous month balance if exists
	var prevBalance int
	var prev domain.Comptime
	errPrev := db.WithContext(ctx).
		Where("customer_id = ? AND service_id = ? AND month < ? AND deleted_at IS NULL",
			customerID, serviceID, monthStr).
		Order("month desc").
		First(&prev).Error
	if errPrev == nil {
		prevBalance = prev.Balance
	}

	comptime = *domain.NewComptime(vendorID, customerID, serviceID, normalizedMonth)
	comptime.Balance = prevBalance

	if err := db.WithContext(ctx).Create(&comptime).Error; err != nil {
		// In case of concurrent creation race, try fetching again
		if errFind := db.WithContext(ctx).
			Where("vendor_id = ? AND customer_id = ? AND service_id = ? AND month = ? AND deleted_at IS NULL",
				vendorID, customerID, serviceID, monthStr).
			First(&comptime).Error; errFind == nil {
			return &comptime, nil
		}
		return nil, err
	}
	return &comptime, nil
}

// GetComptime retrieves a comptime record by ID.
func (a *Repository) GetComptime(ctx context.Context, id int64) (*domain.Comptime, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var comptime domain.Comptime
	err := db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&comptime).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comptime, nil
}

// GetComptimeByCustomerServiceMonth retrieves a comptime record for a given customer, service and month.
func (a *Repository) GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	monthStr := domain.NormalizeMonth(month).Format("2006-01-02")
	var comptime domain.Comptime
	err := db.WithContext(ctx).
		Where("customer_id = ? AND service_id = ? AND month = ? AND deleted_at IS NULL",
			customerID, serviceID, monthStr).
		First(&comptime).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comptime, nil
}

// GetPreviousComptime retrieves the latest comptime record before the given month.
func (a *Repository) GetPreviousComptime(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	monthStr := domain.NormalizeMonth(month).Format("2006-01-02")
	var record domain.Comptime
	err := db.WithContext(ctx).
		Where("customer_id = ? AND service_id = ? AND month < ? AND deleted_at IS NULL",
			customerID, serviceID, monthStr).
		Order("month desc").
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetSubsequentComptimes retrieves all comptime records after the given month ordered chronologically.
func (a *Repository) GetSubsequentComptimes(ctx context.Context, customerID, serviceID int64, month time.Time) ([]domain.Comptime, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	monthStr := domain.NormalizeMonth(month).Format("2006-01-02")
	var records []domain.Comptime
	err := db.WithContext(ctx).
		Where("customer_id = ? AND service_id = ? AND month > ? AND deleted_at IS NULL",
			customerID, serviceID, monthStr).
		Order("month asc").
		Find(&records).Error
	return records, err
}

// SaveComptime persists a comptime record (create or update).
func (a *Repository) SaveComptime(ctx context.Context, comptime *domain.Comptime) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	comptime.UpdatedAt = time.Now()
	if comptime.ID == 0 {
		return db.WithContext(ctx).Create(comptime).Error
	}
	return db.WithContext(ctx).Save(comptime).Error
}

// DeleteComptime soft-deletes a comptime record and its associated child records.
func (a *Repository) DeleteComptime(ctx context.Context, id int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	if err := db.WithContext(ctx).
		Model(&domain.Comptime{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error; err != nil {
		return err
	}
	_ = db.WithContext(ctx).
		Table("planner.comptime_session").
		Where("comptime_id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
	_ = db.WithContext(ctx).
		Table("planner.comptime_invoice_item").
		Where("comptime_id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
	_ = db.WithContext(ctx).
		Table("planner.comptime_extra").
		Where("comptime_id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
	return nil
}

// ListComptime returns paginated comptime records joined with customer and service details.
func (a *Repository) ListComptime(ctx context.Context, filter port.ComptimeFilter) ([]port.ComptimeView, int64, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}

	query := db.WithContext(ctx).
		Table("planner.comptime").
		Select(`
			planner.comptime.id,
			planner.comptime.vendor_id,
			planner.comptime.customer_id,
			planner.customer.nickname as customer_nickname,
			planner.customer.name as customer_name,
			planner.comptime.service_id,
			COALESCE(planner.service.description, planner.service.name) as service_name,
			planner.comptime.month,
			planner.comptime.invoiced_minutes,
			planner.comptime.consumed_minutes,
			(planner.comptime.invoiced_minutes - planner.comptime.consumed_minutes) as month_balance,
			planner.comptime.balance,
			planner.comptime.updated_at
		`).
		Joins("JOIN planner.customer ON planner.customer.id = planner.comptime.customer_id").
		Joins("JOIN planner.service ON planner.service.id = planner.comptime.service_id").
		Where("planner.comptime.deleted_at IS NULL")

	if filter.VendorID > 0 {
		query = query.Where("planner.comptime.vendor_id = ?", filter.VendorID)
	}
	if filter.CustomerID != nil && *filter.CustomerID > 0 {
		query = query.Where("planner.comptime.customer_id = ?", *filter.CustomerID)
	}
	if filter.CustomerNickname != "" {
		query = query.Where("planner.customer.nickname ILIKE ?", "%"+filter.CustomerNickname+"%")
	}
	if filter.ServiceID != nil && *filter.ServiceID > 0 {
		query = query.Where("planner.comptime.service_id = ?", *filter.ServiceID)
	}
	if filter.StartMonth != nil && !filter.StartMonth.IsZero() {
		startStr := domain.NormalizeMonth(*filter.StartMonth).Format("2006-01-02")
		query = query.Where("planner.comptime.month >= ?", startStr)
	}
	if filter.EndMonth != nil && !filter.EndMonth.IsZero() {
		endStr := domain.NormalizeMonth(*filter.EndMonth).Format("2006-01-02")
		query = query.Where("planner.comptime.month <= ?", endStr)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Ordenar pelo mês (decrescente) e pelo cliente (crescente)
	query = query.Order("planner.comptime.month desc, planner.customer.name asc")

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	var results []port.ComptimeView
	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, count, nil
}

// FindComptimeSessionBySessionID finds an active comptime_session by session ID.
func (a *Repository) FindComptimeSessionBySessionID(ctx context.Context, sessionID int64) (*domain.ComptimeSession, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var cs domain.ComptimeSession
	err := db.WithContext(ctx).
		Where("session_id = ? AND deleted_at IS NULL", sessionID).
		First(&cs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// SaveComptimeSession saves or creates a comptime_session.
func (a *Repository) SaveComptimeSession(ctx context.Context, cs *domain.ComptimeSession) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	cs.UpdatedAt = time.Now()
	if cs.ID == 0 {
		return db.WithContext(ctx).Create(cs).Error
	}
	return db.WithContext(ctx).Save(cs).Error
}

// DeleteComptimeSession soft-deletes a comptime_session by session ID.
func (a *Repository) DeleteComptimeSession(ctx context.Context, sessionID int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	return db.WithContext(ctx).
		Model(&domain.ComptimeSession{}).
		Where("session_id = ? AND deleted_at IS NULL", sessionID).
		Update("deleted_at", now).Error
}

// FindComptimeInvoiceItemByInvoiceItemID finds an active comptime_invoice_item by invoice item ID.
func (a *Repository) FindComptimeInvoiceItemByInvoiceItemID(ctx context.Context, invoiceItemID int64) (*domain.ComptimeInvoiceItem, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var cii domain.ComptimeInvoiceItem
	err := db.WithContext(ctx).
		Where("invoice_item_id = ? AND deleted_at IS NULL", invoiceItemID).
		First(&cii).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cii, nil
}

// FindComptimeInvoiceItemsByInvoiceID finds all active comptime_invoice_item records for an invoice ID.
func (a *Repository) FindComptimeInvoiceItemsByInvoiceID(ctx context.Context, invoiceID int64) ([]domain.ComptimeInvoiceItem, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var items []domain.ComptimeInvoiceItem
	err := db.WithContext(ctx).
		Joins("JOIN planner.invoice_item ON planner.invoice_item.id = planner.comptime_invoice_item.invoice_item_id").
		Where("planner.invoice_item.invoice_id = ? AND planner.comptime_invoice_item.deleted_at IS NULL", invoiceID).
		Find(&items).Error
	return items, err
}

// SaveComptimeInvoiceItem saves or creates a comptime_invoice_item.
func (a *Repository) SaveComptimeInvoiceItem(ctx context.Context, cii *domain.ComptimeInvoiceItem) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	cii.UpdatedAt = time.Now()
	if cii.ID == 0 {
		return db.WithContext(ctx).Create(cii).Error
	}
	return db.WithContext(ctx).Save(cii).Error
}

// DeleteComptimeInvoiceItem soft-deletes a comptime_invoice_item by invoice item ID.
func (a *Repository) DeleteComptimeInvoiceItem(ctx context.Context, invoiceItemID int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	return db.WithContext(ctx).
		Model(&domain.ComptimeInvoiceItem{}).
		Where("invoice_item_id = ? AND deleted_at IS NULL", invoiceItemID).
		Update("deleted_at", now).Error
}

// SaveComptimeExtra saves or creates a comptime_extra.
func (a *Repository) SaveComptimeExtra(ctx context.Context, ce *domain.ComptimeExtra) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	ce.UpdatedAt = time.Now()
	if ce.ID == 0 {
		return db.WithContext(ctx).Create(ce).Error
	}
	return db.WithContext(ctx).Save(ce).Error
}

// DeleteComptimeExtra soft-deletes a comptime_extra by ID.
func (a *Repository) DeleteComptimeExtra(ctx context.Context, id int64) error {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	now := time.Now()
	return db.WithContext(ctx).
		Model(&domain.ComptimeExtra{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
}

// FindActiveSessionsForComptime returns all active sessions for a vendor within the specified date range,
// optionally filtered by customer and service. If not specified (nil or <= 0), returns all.
func (a *Repository) FindActiveSessionsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.Session, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	var sessions []domain.Session
	query := db.WithContext(ctx).
		Model(&domain.Session{}).
		Joins("JOIN planner.customer ON planner.customer.id = planner.session.customer_id").
		Where("planner.customer.vendor_id = ?", vendorID).
		Where("planner.session.session_date >= ? AND planner.session.session_date <= ?", startStr, endStr).
		Where("planner.session.deleted_at IS NULL")

	if customerID != nil && *customerID > 0 {
		query = query.Where("planner.session.customer_id = ?", *customerID)
	}
	if serviceID != nil && *serviceID > 0 {
		query = query.Where("planner.session.service_id = ?", *serviceID)
	}

	err := query.Order("planner.session.session_date asc, planner.session.id asc").
		Find(&sessions).Error
	return sessions, err
}

// FindActiveInvoiceItemsForComptime returns all active invoice items for a vendor within the specified period,
// optionally filtered by customer and service. If not specified (nil or <= 0), returns all.
func (a *Repository) FindActiveInvoiceItemsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.InvoiceItem, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	var items []domain.InvoiceItem
	query := db.WithContext(ctx).
		Model(&domain.InvoiceItem{}).
		Joins("JOIN planner.invoice ON planner.invoice.id = planner.invoice_item.invoice_id").
		Joins("JOIN planner.plan ON planner.plan.id = planner.invoice.plan_id").
		Joins("JOIN planner.customer ON planner.customer.id = planner.plan.customer_id").
		Where("planner.customer.vendor_id = ?", vendorID).
		Where("planner.invoice.period_start >= ? AND planner.invoice.period_start <= ?", startStr, endStr).
		Where("planner.invoice_item.deleted_at IS NULL AND planner.invoice.deleted_at IS NULL AND planner.plan.deleted_at IS NULL")

	if customerID != nil && *customerID > 0 {
		query = query.Where("planner.customer.id = ?", *customerID)
	}
	if serviceID != nil && *serviceID > 0 {
		query = query.Where("planner.invoice_item.service_id = ?", *serviceID)
	}

	err := query.Preload("Invoice").
		Preload("Invoice.Plan").
		Preload("Invoice.Plan.Customer").
		Preload("Service").
		Order("planner.invoice.period_start asc, planner.invoice_item.id asc").
		Find(&items).Error
	return items, err
}

// GetInvoiceItemByID retrieves an active invoice item by ID with related Invoice, Plan and Service preloaded.
func (a *Repository) GetInvoiceItemByID(ctx context.Context, id int64) (*domain.InvoiceItem, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var item domain.InvoiceItem
	err := db.WithContext(ctx).
		Preload("Invoice").
		Preload("Invoice.Plan").
		Preload("Invoice.Plan.Customer").
		Preload("Service").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetSessionByID retrieves a session by its ID with Customer and Service preloaded.
func (a *Repository) GetSessionByID(id int64) (*domain.Session, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}
	var session domain.Session
	err := db.Preload("Customer").Preload("Service").Where("id = ? AND deleted_at IS NULL", id).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}


