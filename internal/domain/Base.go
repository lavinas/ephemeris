package domain

type Base struct {
}



// All is a function that returns the domain entity
func All() []interface{} {
	return []interface{}{
		&Client{},
		&Service{},
		&Recurrence{},
		&Price{},
		&Package{},
		&Contract{},
		&Agenda{},
		&Invoice{},
		&InvoiceItem{},
	}
}
