package domain

import (
	"time"
)

// ClientRole represents the role of a client
type ClientRole struct {
	ID   string `gorm:"type:varchar(25); primaryKey"`
	Name string `gorm:"type:varchar(100)"`
}

// ClientContact represents the contact way of a client
type ClientContact struct {
	ID   string `gorm:"type:varchar(25); primaryKey"`
	Name string `gorm:"type:varchar(100)"`
}

// Client represents the client entity
type Client struct {
	ID       string         `gorm:"type:varchar(25); primaryKey"`
	Name     string         `gorm:"type:varchar(100), not null"`
	Document int64          `gorm:"type:numeric(20)"`
	Email    string         `gorm:"type:varchar(100)"`
	Phone    int64          `gorm:"type:numeric(20)"`
	At       time.Time      `gorm:"type:date; not null"`
	Contact  *ClientContact `gorm:"foreignKey:ID"`
	Role     *ClientRole    `gorm:"foreignKey:ID"`
	Bond     *Client        `gorm:"foreignKey:ID"`
}
