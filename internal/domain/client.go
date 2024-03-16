package domain

// ClientRole represents the role of a client
type ClientRole struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100)"`
}

// ClientContactWay represents the contact way of a client
type ClientContactWay struct {
	Base `gorm:"embedded"`
	Name string `gorm:"type:varchar(100)"`
}

// Client represents the client entity
type Client struct {
	Base       `gorm:"embedded"`
	Name       string            `gorm:"type:varchar(100), not null"`
	Document   int64             `gorm:"type:numeric(20), not null; unique"`
	Email      string            `gorm:"type:varchar(100), not null; unique"`
	Phone      int64             `gorm:"type:numeric(20), not null; unique"`
	Role       *ClientRole       `gorm:"foreignKey:ID, not null"`
	ContactWay *ClientContactWay `gorm:"foreignKey:ID, not null"`
	Bond       *Client           `gorm:"foreignKey:ID"`
}
