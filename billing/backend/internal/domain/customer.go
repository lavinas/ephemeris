package domain

import (
	"time"
)

type Customer struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(150);not null"`
	Nickname  string    `gorm:"type:varchar(150)"`
	Status    int       `gorm:"type:int;not null;default:1"`
	Document  *string   `gorm:"type:varchar(50)"`
	Email     *string   `gorm:"type:varchar(150)"`
	Whatsapp  *string   `gorm:"type:varchar(20)"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:now()"`
}

// NewCustomer creates a new Customer instance with the provided details.
func NewCustomer(name, nickname, document, email, whatsapp *string) *Customer {
	return &Customer{
		ID:       0,
		Name:     *name,
		Nickname: *nickname,
		Status:   1,
		Document: document,
		Email:    email,
		Whatsapp: whatsapp,
	}
}

// TableName specifies the table name for Customer model.
func (Customer) TableName() string {
	return "customer"
}
