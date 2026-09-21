package domain

import (
	"time"

	"signup/internal/port"
)

// User represents an individual user within the system.
type User struct {
	ID        int64     `gorm:"primaryKey"`
	VendorID  int64     `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Username  string    `gorm:"not null;unique"`
	PassHash  string    `gorm:"not null"`
	Email     string    `gorm:"null"`
	Whatsapp  string    `gorm:"null"`
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

// GetByUsername retrieves a user by their username.
func (u *User) GetByUsername(repo port.Repository, vendorID int64, username string) (bool, error) {
	conditions := map[string]interface{}{
		"vendor_id": vendorID,
		"username":  username,
	}
	result, err := repo.Find(u, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}
	*u = *(result[0].(*User))
	return true, nil
}

// GetByEmail retrieves a user by their email.
func (u *User) GetByEmail(repo port.Repository, vendorID int64, email string) (bool, error) {
	conditions := map[string]interface{}{
		"vendor_id": vendorID,
		"email":     email,
	}
	result, err := repo.Find(u, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}
	*u = *(result[0].(*User))
	return true, nil
}

// Save persists the user instance to the repository.
func (u *User) Save(repo port.Repository) error {
	return repo.Save(u)
}