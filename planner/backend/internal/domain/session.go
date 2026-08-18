package domain

import (
	"time"
)

// SessionStatus represents string type for status
type SessionStatus string

const (
	StatusDone            SessionStatus = "realizada"
	StatusCancelCharge    SessionStatus = "cancelada_cobrar"
	StatusCancelNotCharge SessionStatus = "cancelada_nao_cobrar"
)

// Session is a struct that dominate session objetcs
type Session struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	CustomerNickname string    `gorm:"not null"`
	SessionDate      time.Time `gorm:"not null"`
	SessionMinutes   int       `gorm:"not null"`
	SessionStatus    string    `gorm:"not null"`
	Comments         *string   `gorm:"type:text"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
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
	}
}

// TableName specifies the table name for Customer model.
func (Session) TableName() string {
	return "session"
}
