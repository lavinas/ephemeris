package domain

// All is a function that returns the domain entity
func All() []interface{} {
	return []interface{}{
		&Client{},
		&Service{},
		&Recurrence{},
		&Price{},
		&Package{},
		&PackageItem{},
		&Contract{},
		&Agenda{},
		&Invoice{},
		&InvoiceItem{},
		&Session{},
	}
}
