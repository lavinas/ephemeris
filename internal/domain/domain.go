package domain

// GetDomain is a function that returns the domain entity
func GetDomain() []interface{} {
	return []interface{}{
		&Client{},
		&Service{},
		&Recurrence{},
		&Price{},
		&Package{},
		&Contract{},
	}
}
