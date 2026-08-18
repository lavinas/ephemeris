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
	SessionStatus    string     `gorm:"not null"`
	Comments         *string    `gorm:"type:text"`
	CreatedAt        time.Time  `gorm:"not null"`
	UpdatedAt        time.Time  `gorm:"not null"`
	DeletedAt        *time.Time `gorm:"index"`
}

// NewSession creates a Session object
func NewSession(nickname string, date time.Time, minutes int, status string, comments *string) *Session {
	return &Session{
		CustomerNickname: nickname,
		SessionDate:      date,
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
func (s *Session) Find(repository port.Repository, page, pagesize int, nickname string, startDate, endDate time.Time,
	minutes int, status string, comments string) ([]Session, error) {
	conditions := map[string]interface{}{}
	conditions["deleted_at IS NULL"] = nil
	if nickname != "" {
		conditions["customer_nickname = "] = nickname
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
	if status != "" {
		conditions["session_status = ?"] = status
	}
	if comments != "" {
		conditions["comments like ?"] = "%" + comments + "%"
	}
	results, err := repository.Find(page, pagesize, conditions)
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, len(results))
	for i, v := range results {
		sessions[i] = v.(Session)
	}
	return sessions, nil
}

// Get is a helper function to get a record from the database
func (s *Session) Get(repository port.Repository, id int64) (bool, error) {
	conditions := map[string]interface{}{
		"id = ?":          id,
		"deleted_at IS NULL": nil,
	}
	results, err := repository.Find(1, 1, conditions)
	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		return false, nil
	}
	session := results[0].(Session)
	s.ID = session.ID
	s.CustomerNickname = session.CustomerNickname
	s.SessionDate = session.SessionDate
	s.SessionMinutes = session.SessionMinutes
	s.SessionStatus = session.SessionStatus
	return true, nil
}
