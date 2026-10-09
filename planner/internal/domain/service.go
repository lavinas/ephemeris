package domain

import (
	"time"
)

// Service represents the service domain model.
type Service struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`
	VendorID       int64     `gorm:"not null;index"`
	Name           string    `gorm:"not null"`
	Description    *string   `gorm:"type:text"`
	SessionMinutes *int      `gorm:"column:session_minutes"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

// NewService creates a new Service instance with the provided details.
func NewService(vendorID int64, name string, description *string, sessionMinutes *int) *Service {
	return &Service{
		VendorID:       vendorID,
		Name:           name,
		Description:    description,
		SessionMinutes: sessionMinutes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// TableName specifies the table name for Service model.
func (Service) TableName() string {
	return "service"
}
