package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"planner/internal/domain"
)

// GetSessionByID retrieves a session by its ID.
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

// FindMatchingInvoiceItem finds an active invoice item matching the session criteria:
// - Same customer (plan.customer_id)
// - Same service (invoice_item.service_id)
// - sessionDate within [invoice.period_start, invoice.period_end]
// - Not deleted
func (a *Repository) FindMatchingInvoiceItem(customerID, serviceID int64, sessionDate time.Time) (*domain.InvoiceItem, *domain.Invoice, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}

	dateStr := sessionDate.Format("2006-01-02")

	var item domain.InvoiceItem
	err := db.Model(&domain.InvoiceItem{}).
		Joins("JOIN planner.invoice ON planner.invoice.id = planner.invoice_item.invoice_id").
		Joins("JOIN planner.plan ON planner.plan.id = planner.invoice.plan_id").
		Where("planner.plan.customer_id = ? AND planner.invoice_item.service_id = ?", customerID, serviceID).
		Where("planner.invoice.period_start <= ? AND planner.invoice.period_end >= ?", dateStr, dateStr).
		Where("planner.invoice.deleted_at IS NULL AND planner.invoice_item.deleted_at IS NULL AND planner.plan.deleted_at IS NULL").
		Preload("Invoice").
		Preload("Invoice.Plan").
		Preload("Service").
		Order("planner.invoice.invoicing_date asc, planner.invoice.id asc").
		First(&item).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	return &item, item.Invoice, nil
}

// IsSessionReconciled checks if a session is already linked to an active invoice item and invoice.
func (a *Repository) IsSessionReconciled(sessionID int64) (bool, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}

	var count int64
	err := db.Model(&domain.InvoiceItemSession{}).
		Joins("JOIN planner.invoice_item ON planner.invoice_item.id = planner.invoice_item_session.invoice_item_id").
		Joins("JOIN planner.invoice ON planner.invoice.id = planner.invoice_item.invoice_id").
		Where("planner.invoice_item_session.session_id = ?", sessionID).
		Where("planner.invoice_item.deleted_at IS NULL AND planner.invoice.deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindUnreconciledSessions finds billable sessions for a customer and service within a date range
// that are not yet associated with any active invoice.
func (a *Repository) FindUnreconciledSessions(customerID, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	subQuery := db.Table("planner.invoice_item_session").
		Select("planner.invoice_item_session.session_id").
		Joins("JOIN planner.invoice_item ON planner.invoice_item.id = planner.invoice_item_session.invoice_item_id").
		Joins("JOIN planner.invoice ON planner.invoice.id = planner.invoice_item.invoice_id").
		Where("planner.invoice_item.deleted_at IS NULL AND planner.invoice.deleted_at IS NULL")

	var sessions []domain.Session
	err := db.Model(&domain.Session{}).
		Where("planner.session.customer_id = ? AND planner.session.service_id = ?", customerID, serviceID).
		Where("planner.session.session_date >= ? AND planner.session.session_date <= ?", startStr, endStr).
		Where("planner.session.deleted_at IS NULL").
		Where("planner.session.session_status NOT IN (?)", []string{string(domain.StatusCancelNotCharge), "Cancelada/Não Cobrar"}).
		Where("planner.session.id NOT IN (?)", subQuery).
		Order("planner.session.session_date asc, planner.session.id asc").
		Find(&sessions).Error

	return sessions, err
}

// AssociateSessionToInvoiceItem creates the link in invoice_item_session and increments realized_quantity by 1.
func (a *Repository) AssociateSessionToInvoiceItem(invoiceItemID, sessionID int64) error {
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

	// Idempotency: verify if association already exists
	var existingCount int64
	if err := tx.Model(&domain.InvoiceItemSession{}).
		Where("invoice_item_id = ? AND session_id = ?", invoiceItemID, sessionID).
		Count(&existingCount).Error; err != nil {
		tx.Rollback()
		return err
	}
	if existingCount > 0 {
		tx.Rollback()
		return nil
	}

	// Insert association
	assoc := domain.NewInvoiceItemSession(invoiceItemID, sessionID)
	assoc.CreatedAt = now
	if err := tx.Omit("InvoiceItem", "Session").Create(assoc).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Increment realized_quantity
	if err := tx.Model(&domain.InvoiceItem{}).
		Where("id = ?", invoiceItemID).
		Updates(map[string]interface{}{
			"realized_quantity": gorm.Expr("realized_quantity + 1"),
			"updated_at":        now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DissociateSessionFromInvoiceItem removes the association from invoice_item_session and decrements realized_quantity.
func (a *Repository) DissociateSessionFromInvoiceItem(sessionID int64) (*domain.InvoiceItem, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}

	now := time.Now()
	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var assoc domain.InvoiceItemSession
	err := tx.Where("session_id = ?", sessionID).First(&assoc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, nil // Nothing to dissociate
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Delete association
	if err := tx.Delete(&assoc).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Load and decrement invoice item realized_quantity
	var item domain.InvoiceItem
	if err := tx.Where("id = ?", assoc.InvoiceItemID).First(&item).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	newQty := item.RealizedQuantity - 1
	if newQty < 0 {
		newQty = 0
	}

	if err := tx.Model(&item).Updates(map[string]interface{}{
		"realized_quantity": newQty,
		"updated_at":        now,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	item.RealizedQuantity = newQty
	return &item, tx.Commit().Error
}

// FindActivePayPerViewPlan finds an active pay-per-view plan and the specific plan_item for the given customer and service.
func (a *Repository) FindActivePayPerViewPlan(customerID, serviceID int64, sessionDate time.Time) (*domain.Plan, *domain.PlanItem, error) {
	db := a.DB
	if a.Tx != nil {
		db = a.Tx
	}

	dateStr := sessionDate.Format("2006-01-02")

	var plan domain.Plan
	err := db.Model(&domain.Plan{}).
		Joins("JOIN planner.plan_item ON planner.plan_item.plan_id = planner.plan.id").
		Where("planner.plan.customer_id = ? AND planner.plan.plan_type = ?", customerID, int(domain.PlanTypePayPerView)).
		Where("planner.plan_item.service_id = ? AND planner.plan_item.deleted_at IS NULL", serviceID).
		Where("planner.plan.plan_start <= ?", dateStr).
		Where("(planner.plan.plan_end IS NULL OR planner.plan.plan_end >= ?)", dateStr).
		Where("planner.plan.deleted_at IS NULL").
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL")
		}).
		Preload("Items.Service").
		Preload("PayPerView").
		Preload("Customer").
		Order("planner.plan.plan_start desc, planner.plan.id desc").
		First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	var matchedItem *domain.PlanItem
	for i := range plan.Items {
		if plan.Items[i].ServiceID == serviceID {
			matchedItem = &plan.Items[i]
			break
		}
	}

	return &plan, matchedItem, nil
}
