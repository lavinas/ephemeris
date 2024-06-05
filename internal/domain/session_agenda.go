package domain

var (
	statusSessionAgenda = []string{}

)


// SessionAgenda represents the domain for a session linked to agenda
type SessionAgenda struct {
	SessionID string `gorm:"type:varchar(150); primaryKey"`
	AgendaID  string `gorm:"type:varchar(150); primaryKey"`
	StatusID  string `gorm:"type:varchar(50); not null; index"`
}