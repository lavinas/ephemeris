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
	ID               int64      `gorm:"primaryKey;autoIncrement"`
	CustomerNickname string     `gorm:"not null"`
	SessionDate      time.Time  `gorm:"not null"`
	SessionMinutes   int        `gorm:"not null"`
	SessionService   string     `gorm:"not null"`
	SessionStatus    string     `gorm:"not null"`
	Comments         *string    `gorm:"type:text"`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
	DeletedAt        *time.Time `gorm:"index"`
}

// NewSession creates a Session object
func NewSession(nickname string, date time.Time, minutes int, service string, status string, comments *string) *Session {
	return &Session{
		CustomerNickname: nickname,
		SessionDate:      date,
		SessionService:   service,
		SessionStatus:    status,
		SessionMinutes:   minutes,
		Comments:         comments,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
	}
}

// TableName specifies the table name for Customer model.
func (Session) TableName() string {
	return "session"
}

// Find is a helper function to find records in the database
func (s *Session) Find(repository SessionRepository, page, pagesize int, id int64, nickname string, startDate, endDate time.Time,
	minutes int, service, status string, comments string) ([]Session, int64, error) {
	conditions := map[string]interface{}{}
	conditions["deleted_at IS NULL"] = nil
	if id != 0 {
		conditions["id = ?"] = id
	}
	if nickname != "" {
		conditions["customer_nickname like ?"] = "%" + nickname + "%"
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
	if service != "" {
		conditions["session_service = ?"] = service
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

// FindUsers is a helper function to find session users in the database based on conditions
func (s *Session) FindUsers(repository SessionRepository, startDate, endDate time.Time,
	minutes int, service, status string) ([]string, error) {
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
	if service != "" {
		conditions["session_service = ?"] = service
	}
	if status != "" {
		conditions["session_status = ?"] = status
	}
	results, err := repository.FindGroup(conditions, "customer_nickname")
	if err != nil {
		return nil, err
	}
	users := []string{}
	for _, v := range results {
		users = append(users, v["customer_nickname"].(string))
	}
	return users, nil
}
