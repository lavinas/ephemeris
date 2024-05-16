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
		&PackageAppend{},
		&PriceCrud{},
		&RecurrenceCrud{},
		&ServiceCrud{},
		&SessionCrud{},
	}
}
