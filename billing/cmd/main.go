package main

import (
	"context"

	"billing/internal/adapter"
	"billing/internal/service"
	"billing/internal/dto"
	"fmt"
)

func main() {
	fmt.Println("Starting Billing Service...")
	// Initialize the logger
	logger, err := adapter.NewSimpleLogger("stdout", 1)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return

	}
	defer logger.Close()
	// Initialize Config
	cfg, err := adapter.NewConfig("billing.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	ctx := context.Background()
	repo, err := adapter.NewPostgresRepository(host, user, password, dbname, sslmode, timezone, port, timeout, schema, &ctx)
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
		InvoiceItems:     []dto.CreateInvoiceItemDTO{*invoiceItemDTO},
	}
	response := service.CreateInvoice(*invoiceDTO)
	fmt.Printf("Create Invoice Response: %+v\n", response)
}
