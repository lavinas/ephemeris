package domain

import (
	"time"
)

// Base represents the base entity of all entities
type Base struct {
	ID        string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt time.Time `gorm:"type:datetime; not null"`
}
