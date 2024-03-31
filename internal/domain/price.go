package domain

import (
	"time"
)

// Price represents the price entity
type Price struct {
	ID        string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt time.Time `gorm:"type:datetime; not null"`
	Unit      float64   `gorm:"type:numeric(20,2); not null"`
	Pack      float64   `gorm:"type:numeric(20,2); not null"`
}
