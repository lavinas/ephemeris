package domain

// Service represents the service entity
type Service struct {
	Base     `gorm:"embedded"`
	ID       string `gorm:"type:varchar(25); primaryKey"`
	Name     string `gorm:"type:varchar(100), not null"`
	Duration int64  `gorm:"type:numeric(20), not null"`
}
