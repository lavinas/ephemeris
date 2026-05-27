package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/dto"
	"billing/internal/service"
	"fmt"
	"encoding/json"
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

	customer1 := dto.CreateCustomerRequestItem{
		Name:     "John Doe",
		Nickname: "Johnny",
		Document: "123456789",
		Email:    "john.doe@example.com",
		Whatsapp: "(11)91234-5678",
	}
	customer2 := dto.CreateCustomerRequestItem{
		Name:     "John Doe",
		Nickname: "Johnny1",
		Document: "1234567891",
		Email:    "john.doe@example.com",
		Whatsapp: "(11)91234-5678",
	}

	response := service.Run(&dto.CreateCustomerRequest{Items: []dto.CreateCustomerRequestItem{customer1, customer2}})

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling response: %v\n", err)
		return
	}
	fmt.Printf("Create Customer Response: %s\n", responseJSON)

}
