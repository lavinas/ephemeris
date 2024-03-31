package domain

import (
	"time"
)

// Service represents the service entity
type Service struct {
	ID          string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt   time.Time `gorm:"type:datetime; not null"`
	Name     string `gorm:"type:varchar(100), not null"`
	Duration int64  `gorm:"type:numeric(20), not null"`
}
