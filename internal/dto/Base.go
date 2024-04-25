package dto

func All() []interface{} {
	return []interface{}{
		&AgendaCrud{},
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
