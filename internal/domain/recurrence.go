package domain

import (
	"time"
)

// Cycle represents the cycle entity
type Cycle struct {
	ID        string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt time.Time `gorm:"type:datetime; not null"`
	Name      string    `gorm:"type:varchar(100), not null"`
}

// Recurrence represents the recurrence entity
type Recurrence struct {
	ID        string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt time.Time `gorm:"type:datetime; not null"`
	Name      string    `gorm:"type:varchar(100), not null"`
	Cycle     *Cycle    `gorm:"foreignKey:ID, not null"`
	Quantity  int64     `gorm:"type:numeric(20), not null"`
	Limit     int64     `gorm:"type:numeric(20), not null"`
}
