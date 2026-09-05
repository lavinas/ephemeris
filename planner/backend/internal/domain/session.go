package domain

import (
	"time"

	"planner/internal/port"
)

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
func (s *Session) Find(repository port.Repository, page, pagesize int, id int64, nickname string, startDate, endDate time.Time,
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
	orders := []string{"session_date desc", "customer_nickname"}
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
func (s *Session) FindUsers(repository port.Repository, startDate, endDate time.Time,
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
