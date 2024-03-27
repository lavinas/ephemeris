package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	ErrEmptyID   = "id is empty"
	ErrInvalidID = "id is invalid. It should not have spaces"
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
	if len(strings.Split(b.ID, " ")) > 1 {
		return errors.New(ErrInvalidID)
	}
	return nil
}

// Formmat formats the base entity
func (b *Base) Format() {
	if err := b.Validate(); err != nil {
		b.ID = ""
		return
	}
	b.ID = strings.TrimSpace(b.ID)
	b.ID = strings.ToLower(b.ID)
}

// GetBase returns a new base object
func GetDomain() []interface{} {
	return []interface{}{
		&Client{},
	}
}
