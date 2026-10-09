package domain

import (
	"time"
)

// SessionRepository defines the minimal repository methods needed by Session domain.
type SessionRepository interface {
	Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error)
	FindCount(conditions map[string]interface{}) (int64, error)
	FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error)
}

// SessionStatus represents string type for status
type SessionStatus string

const (
	StatusDone            SessionStatus = "realizada"
	StatusCancelCharge    SessionStatus = "cancelada_cobrar"
	StatusCancelNotCharge SessionStatus = "cancelada_nao_cobrar"

	ErrSessionNotFound = "session not found"
)

// Session is a struct that dominate session objetcs
type Session struct {
	ID             int64      `gorm:"primaryKey;autoIncrement"`
	CustomerID     int64      `gorm:"not null;index"`
	Customer       *Customer  `gorm:"foreignKey:CustomerID;references:ID"`
	ServiceID      int64      `gorm:"not null;index"`
	Service        *Service   `gorm:"foreignKey:ServiceID;references:ID"`
	SessionDate    time.Time  `gorm:"not null"`
	SessionMinutes int        `gorm:"not null"`
	SessionStatus  string     `gorm:"not null"`
	Comments       *string    `gorm:"type:text"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
	DeletedAt      *time.Time `gorm:"index"`
}

// NewSession creates a Session object
func NewSession(customerID, serviceID int64, date time.Time, minutes int, status string, comments *string) *Session {
	return &Session{
		CustomerID:     customerID,
		ServiceID:      serviceID,
		SessionDate:    date,
		SessionStatus:  status,
		SessionMinutes: minutes,
		Comments:       comments,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
	}
}

// TableName specifies the table name for Session model.
func (Session) TableName() string {
	return "session"
}

// Find is a helper function to find records in the database
func (s *Session) Find(repository SessionRepository, page, pagesize int, id int64, customerIDs []int64, startDate, endDate time.Time,
	minutes int, serviceID int64, status string, comments string) ([]Session, int64, error) {
	conditions := map[string]interface{}{}
	conditions["deleted_at IS NULL"] = nil
	if id != 0 {
		conditions["id = ?"] = id
	}
	if len(customerIDs) == 1 {
		conditions["customer_id = ?"] = customerIDs[0]
	} else if len(customerIDs) > 1 {
		conditions["customer_id IN (?)"] = customerIDs
	}
	if !startDate.IsZero() {
		conditions["session_date >= ?"] = startDate
	}
	if !endDate.IsZero() {
		conditions["session_date <= ?"] = endDate
	}
	if minutes != 0 {
		conditions["session_minutes = ?"] = minutes
	}
	if serviceID != 0 {
		conditions["service_id = ?"] = serviceID
	}
	if status != "" {
		conditions["session_status = ?"] = status
	}
	if comments != "" {
		conditions["comments like ?"] = "%" + comments + "%"
	}
	var count int64
	count, err := repository.FindCount(conditions)
	if err != nil {
		return nil, 0, err
	}
	orders := []string{"session_date desc", "id desc"}
	results, err := repository.Find(page, pagesize, conditions, orders...)
	if err != nil {
		return nil, 0, err
	}
	sessions := make([]Session, len(results))
	for i, v := range results {
		sessions[i] = v.(Session)
	}
	return sessions, count, nil
}

// FindCustomerIDs is a helper function to find distinct session customer IDs in the database based on conditions
func (s *Session) FindCustomerIDs(repository SessionRepository, startDate, endDate time.Time,
	minutes int, serviceID int64, status string) ([]int64, error) {
	conditions := map[string]interface{}{}
	conditions["deleted_at IS NULL"] = nil
	if !startDate.IsZero() {
		conditions["session_date >= ?"] = startDate
	}
	if !endDate.IsZero() {
		conditions["session_date <= ?"] = endDate
	}
	if minutes != 0 {
		conditions["session_minutes = ?"] = minutes
	}
	if serviceID != 0 {
		conditions["service_id = ?"] = serviceID
	}
	if status != "" {
		conditions["session_status = ?"] = status
	}
	results, err := repository.FindGroup(conditions, "customer_id")
	if err != nil {
		return nil, err
	}
	ids := []int64{}
	for _, v := range results {
		if idVal, ok := v["customer_id"]; ok {
			switch id := idVal.(type) {
			case int64:
				ids = append(ids, id)
			case int:
				ids = append(ids, int64(id))
			case float64:
				ids = append(ids, int64(id))
			}
		}
	}
	return ids, nil
}
