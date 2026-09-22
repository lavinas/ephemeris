package domain

import (
	"fmt"
	"time"

	"signup/internal/port"
)

// User represents an individual user within the system.
type User struct {
	DomainBase
	ID        int64     `gorm:"primaryKey"`
	VendorID  int64     `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Username  string    `gorm:"not null;unique"`
	PassHash  string    `gorm:"not null"`
	Email     *string   `gorm:"null"`
	Whatsapp  *string   `gorm:"null"`
	Status    *int      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// NewUser represents the input data required to create a new user.
func NewUser(repo port.Repository, vendorID int64, name, username, passHash string, email, whatsapp *string) *User {
	status := 1
	return &User{
		DomainBase: DomainBase{Repo: repo},
		VendorID:   vendorID,
		Name:       name,
		Username:   username,
		Email:      email,
		PassHash:   passHash,
		Whatsapp:   whatsapp,
		Status:     &status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// StartUser represents the input data required to create a new user.
func StartUser(repo port.Repository) *User {
	return &User{DomainBase: DomainBase{Repo: repo}}
}

// TableName specifies the table name for User model.
func (User) TableName() string {
	return "users"
}

// GetByUsername retrieves a user by their username.
func (u *User) GetByUsername(vendorID int64, username string) (bool, error) {
	conditions := map[string]interface{}{
		"vendor_id": vendorID,
		"username":  username,
	}
	result, err := u.Repo.Find(u, conditions, 1, 1)
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
func (u *User) GetByEmail(vendorID int64, email string) (bool, error) {
	conditions := map[string]interface{}{
		"vendor_id": vendorID,
		"email":     email,
	}
	result, err := u.Repo.Find(u, conditions, 1, 1)
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}
	*u = *(result[0].(*User))
	return true, nil
}

func (u *User) Validate() error {
	if u.VendorID <= 0 {
		return fmt.Errorf("vendor ID is required")
	}
	vendor := Vendor{}
	if _, err := vendor.GetByID(u.VendorID); err != nil {
		return fmt.Errorf("vendor not found")
	}
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	if u.Username == "" {
		return fmt.Errorf("nickname is required")
	}
	if u.Username != "" {
		if err := u.ValidateNickname(u.Username); err != nil {
			return fmt.Errorf("nickname is invalid")
		}
	}
	if u.PassHash == "" {
		return fmt.Errorf("password is required")
	}
	if u.Email != nil {
		if err := u.ValidateEmail(*u.Email); err != nil {
			return fmt.Errorf("email is invalid")
		}
	}
	if u.Whatsapp != nil {
		if _, err := u.ValidateCellNumber(*u.Whatsapp); err != nil {
			return fmt.Errorf("whatsapp is invalid")
		}
	}
	if u.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}
	if u.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	if u.Status == nil {
		status := 1
		u.Status = &status
	}
	if *u.Status != 1 && *u.Status != 0 {
		return fmt.Errorf("status is invalid")
	}
	return nil
}

func (u *User) Find() ([]port.Domain, error) {
	conditions := map[string]interface{}{}
	if u.VendorID > 0 {
		conditions["vendor_id = ?"] = u.VendorID
	}
	if u.Name != "" {
		conditions["name like ?"] = "%" + u.Name + "%"
	}
	if u.Username != "" {
		conditions["username like ?"] = "%" + u.Username + "%"
	}
	if u.Email != nil {
		conditions["email like ?"] = "%" + *u.Email + "%"
	}
	if u.Whatsapp != nil {
		conditions["whatsapp like ?"] = "%" + *u.Whatsapp + "%"
	}
	if u.Status != nil {
		conditions["status = ?"] = *u.Status
	}
	result, err := u.Repo.Find(u, conditions, 0, 0)
	if err != nil {
		return nil, err
	}
	out := make([]port.Domain, len(result))
	for i, r := range result {
		out[i] = r.(*User)
	}
	return out, nil
}

// Save persists the user instance to the repository.
func (u *User) Save() error {
	return u.Repo.Save(u)
}
