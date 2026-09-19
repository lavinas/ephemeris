package domain

import "time"
// User represents an individual user within the system.
type User struct {
	ID       int64  `gorm:"primaryKey"`
	VendorID int64  `gorm:"not null"`
	Name     string `gorm:"not null"`
	Username string `gorm:"not null;unique"`
	PassHash string `gorm:"not null"`
	Email    string `gorm:"null"`
	Whatsapp  string `gorm:"null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// NewUser represents the input data required to create a new user.
func NewUser(vendorID int64, name, username, email, passHash, whatsapp string) *User {
	return &User{
		VendorID: vendorID,
		Name:     name,
		Username: username,
		Email:    email,
		PassHash: passHash,
		Whatsapp: whatsapp,
	}
}

// TableName specifies the table name for User model.
func (User) TableName() string {
	return "users"
}