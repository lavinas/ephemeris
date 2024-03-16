package domain

// Cycle represents the cycle entity
type Cycle struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100), not null"`
}

// Recurrence represents the recurrence entity
type Recurrence struct {
	Base     `gorm:"embedded"`
	Name     string `gorm:"type:varchar(100), not null"`
	Cycle    *Cycle `gorm:"foreignKey:ID, not null"`
	Quantity int64  `gorm:"type:numeric(20), not null"`
	Limit    int64  `gorm:"type:numeric(20), not null"`
}
