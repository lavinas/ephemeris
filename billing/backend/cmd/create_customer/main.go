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
	repo, err := driven.NewPostgresRepository(host, user, password, dbname, sslmode,
		timezone, port, timeout, schema)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}
	defer repo.Close()

	// Initialize Service
	service := service.NewCreateCustomerService(repo, logger)

	customer := dto.CreateCustomerRequest{
		Name:     "John Doe",
		Nickname: "Johnny",
		Document: "123456789",
		Email:    "john.doe@example.com",
		Whatsapp: "+1234567890",
	}
	response := service.Run([]dto.CreateCustomerRequest{customer})
	fmt.Printf("Create Customer Response: %+v\n", response)

	// Code here to initialize the service and run the create customer logic, e.g.:
	//

}
