package dto

func All() []interface{} {
	return []interface{}{
		&AgendaCrud{},
		&AgendaMake{},
		&ClientCrud{},
		&ContractCrud{},
		&InvoiceCrud{},
		&InvoiceItemCrud{},
		&PackageCrud{},
		&PriceCrud{},
		&RecurrenceCrud{},
		&ServiceCrud{},
	}
}
