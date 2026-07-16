package main

import (
	"billing/internal/adapter/driven"
	"billing/internal/adapter/driver"

	"fmt"
	"os"
)

// Main function to initialize the API server
func main() {
	// Initialize Config
	cfg, err := driven.NewConfig("billing.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	// Initialize the logger
	logOutput, logLevel := cfg.GetLogData()
	logger, err := driven.NewSimpleLogger(logOutput, logLevel)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}
	defer logger.Close()
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := driven.NewPostgresRepository(host, user, password, dbname, sslmode,
		timezone, port, timeout, schema)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}
	defer repo.Close()
	// Issuer initialization
	issuer := driven.NewIssuerFile(cfg.GetIssuerFilePath(), cfg.GetIssuerFilePattern(), logger)
	// Biller initialization
	biller := driven.NewBillerMaroto(logger)
	// Pixer initialization
	pixer := driven.NewPixer(logger)
	// Initialize API Handler
	os.Setenv("TZ", timezone)
	logger.IPrintf(0, "starting API server on :8080")
	apiHandler := driver.NewAPIHandler(":8080", logger, repo, issuer, biller, pixer)
	apiHandler.Run(":8080")
	logger.IPrintf(0, "logger and database closed 2")
}
