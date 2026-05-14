package main

import (
	"context"

	"billing/internal/adapter"
	"billing/internal/service"
	"fmt"
)

func main() {
	// Initialize the logger
	logger, err := adapter.NewSimpleLogger("stdout", 1)
	if err != nil {
		panic(fmt.Errorf("Error initializing logger: %v", err))
		
	}
	defer logger.Close()
	// Initialize Config
	cfg, err := adapter.NewConfig("config.json")
	if err != nil {
		panic(fmt.Errorf("Error loading config: %v", err))
	}
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, connect_timeout := cfg.GetDBData()
	ctx := context.Background()
	repo, err := adapter.NewPostgresRepository(host, user, password, dbname, sslmode, timezone, port, connect_timeout, &ctx)
	if err != nil {
		panic(fmt.Errorf("Error initializing repository: %v", err))
	}
	defer repo.Close()
	// Initialize Service
	service := service.NewCreateInvoiceService(repo, logger)

	fmt.Println("Service initialized:", service)
}
