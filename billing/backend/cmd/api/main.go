package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/adapter/driver"

	"fmt"
)

func main() {
	// Initialize the logger
	logger, err := driven.NewSimpleLogger("stdout", 0)
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
	// Initialize API Handler
	logger.IPrintf(0, "starting API server on :8080")
	apiHandler := driver.NewAPIHandler(":8080", logger, repo)
	apiHandler.Run(":8080")
	logger.IPrintf(0, "logger and database closed")
}
