package domain

// Price represents the price entity
type Price struct {
	Base `gorm:"embedded"`
	Unit float64 `gorm:"type:numeric(20,2); not null"`
	Pack float64 `gorm:"type:numeric(20,2); not null"`
}
