package main

import (
	"context"

	"billing/internal/adapter/driven"
	"billing/internal/adapter/driver"
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
	ctx := context.Background()
	repo, err := driven.NewPostgresRepository(host, user, password, dbname, sslmode, timezone, port, timeout, schema, &ctx)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}
	defer repo.Close()
	// Initialize Service
	service := service.NewCreateCustomerService(repo, logger)

	// Initialize Driver
	driver := driver.NewCreateCustomerCsv(*service)

	// Run the driver
	if err := driver.Run(); err != nil {
		fmt.Printf("Error running driver: %v\n", err)
		return
	}
}
