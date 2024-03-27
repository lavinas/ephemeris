package domain

import (
	"time"
	"errors"
)

const (
	ErrEmptyID = "id is empty"
)

// Base represents the base entity of all entities
type Base struct {
	ID        string    `gorm:"type:varchar(25); primaryKey"`
	CreatedAt time.Time `gorm:"type:datetime; not null"`
}

// NewBase creates a new base
func NewBase(id string) *Base {
	return &Base{
		ID:        id,
		CreatedAt: time.Now(),
	}
}

// Validate validates the base entity
func (b *Base) Validate() error {
	if b.ID == "" {
		return errors.New(ErrEmptyID)
	}
	return nil
}

// GetBase returns a new base object
func GetDomain() []interface{} {
	return []interface{}{
		&Client{},
	}
}
