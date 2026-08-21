package main

import (
	"fmt"
	"os"

	"planner/internal/adapter/config"
	"planner/internal/adapter/logger"
	"planner/internal/adapter/repository"
	"planner/internal/adapter/http"
)

const (
	addr = ":8083"
	configFile = "planner.json"
)

// Main function to initialize the API server
func main() {
	// Initialize Config
	logger, repo, handler, err := startAll()
	if err != nil {
		fmt.Printf("Error initializing components: %v\n", err)
		return
	}
	defer repo.Close()
	defer logger.Close()
	if err := handler.Run(addr); err != nil {
		fmt.Printf("Error running server: %v\n", err)
	}

}

// initComponents initializes the logger and repository components
func startAll() (*logger.Logger, *repository.Repository, *http.Handler, error) {
	// Initialize Config
	cfg, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return nil, nil, nil, fmt.Errorf("error loading config: %v", err)
	}
	// Initialize the logger
	logOutput, logLevel := cfg.GetLogData()
	logger, err := logger.NewLogger(logOutput, logLevel)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error initializing logger: %v", err)
	}
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := repository.NewRepository(host, user, password, dbname, sslmode,
		timezone, port, timeout, schema)
	if err != nil {
		logger.Close()
		return nil, nil, nil, fmt.Errorf("error initializing repository: %v", err)
	}
	// Set the timezone environment variable
	os.Setenv("TZ", timezone)
	return logger, repo, http.NewHandler(repo, logger), nil
}




