package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"billing/internal/service"
	"fmt"
)

func main() {
	// Initialize the logger
	logger, err := driven.NewSimpleLogger("stdout", 1)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return

	}
	defer logger.Close()
	// Initialize Config
	cfg, err := driven.NewConfig("billing.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := driven.NewPostgresRepository(host, user, password, dbname, sslmode, timezone,
		port, timeout, schema)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}
	defer repo.Close()
	// Initialize Service
	service := service.NewCreateInvoiceService(repo, logger)

	invoiceItemDTO := &dto.CreateInvoiceItemDTO{
		Description: "Sample Item",
		Quantity:    2,
		UnitPrice:   50.00,
	}

	invoiceDTO := &dto.CreateInvoiceRequest{
		CustomerName:     "John Doe",
		CustomerEmail:    "john.doe@example.com",
		CustomerWhatsapp: "+1234567890",
		CustomerDocument: "123456789",
		Notes:            "Sample invoice",
		InvoiceItems:     []dto.CreateInvoiceItemDTO{*invoiceItemDTO, *invoiceItemDTO},
	}
	response := service.Run(invoiceDTO)
	fmt.Printf("Create Invoice Response: %+v\n", response)
}
