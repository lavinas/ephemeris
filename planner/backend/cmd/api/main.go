package main

import (
	"planner/internal/adapter"

	"fmt"
	"os"
)

// Main function to initialize the API server
func main() {
	// Initialize Config
	cfg, err := adapter.NewConfig("planner.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	// Initialize the logger
	logOutput, logLevel := cfg.GetLogData()
	logger, err := adapter.NewLogger(logOutput, logLevel)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}
	defer logger.Close()
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := adapter.NewRepository(host, user, password, dbname, sslmode,
		timezone, port, timeout, schema)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}
	defer repo.Close()
	os.Setenv("TZ", timezone)
	logger.IPrintf(0, "starting API server on :8083")
	apiHandler := adapter.NewHandler(":8083", logger, repo)
	apiHandler.Run(":8083")
	logger.IPrintf(0, "logger and database closed 2")
}
