package domain

// GetDomain is a function that returns the domain entity
func GetDomain() []interface{} {
	return []interface{}{
		&Client{},
		&ClientRole{},
		&Service{},
		&Recurrence{},
	}
}
