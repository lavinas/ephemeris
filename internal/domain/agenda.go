package domain

import (
	"time"
)

// AgendaType represents the agenda type entity
type AgendaType struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

// AgendaStatus represents the agenda status entity
type AgendaStatus struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

// Agenda represents the agenda entity
type Agenda struct {
	Base         `gorm:"embedded"`
	Contract     *Contract     `gorm:"foreignKey:ID, not null"`
	Start        time.Time     `gorm:"type:datetime; not null"`
	End          time.Time     `gorm:"type:datetime; not null"`
	Type         *AgendaType   `gorm:"foreignKey:ID, not null"`
	Status       *AgendaStatus `gorm:"foreignKey:ID, not null"`
	Bond         *Agenda       `gorm:"foreignKey:ID"`
	BillingMonth int64         `gorm:"type:numeric(20)"`
}
