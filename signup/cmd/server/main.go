package main

import (
	"fmt"
	"os"

	"signup/internal/adapter/config"
	"signup/internal/adapter/http"
	"signup/internal/adapter/logger"
	"signup/internal/adapter/messaging"
	"signup/internal/adapter/repository"
	"signup/internal/port"
)

const (
	configFile = "signup.json"
)

// Main function to initialize the API server
func main() {
	// Initialize Config
	logger, repo, publisher, handler, cfg, err := startAll()
	if err != nil {
		fmt.Printf("Error initializing components: %v\n", err)
		return
	}
	defer func() {
		logger.IPrintf(0, "Logger closed")
		logger.Close()
	}()
	defer func() {
		repo.Close()
		logger.IPrintf(0, "Repository closed")
	}()
	defer func() {
		publisher.Close()
		logger.IPrintf(0, "Publisher closed")
	}()
	if err := handler.Run(cfg.GetWebAddr()); err != nil {
		logger.IPrintf(0, "Error running server: %v", err)
	}
}

// startAll initializes the logger, repository, publisher, and handler components
func startAll() (*logger.Logger2, *repository.Repository, port.CustomerEventPublisher, *http.Handler, *config.Config, error) {
	// Initialize Config
	cfg, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return nil, nil, nil, nil, nil, fmt.Errorf("error loading config: %v", err)
	}
	// Initialize the logger
	logOutput, logLevel := cfg.GetLogData()
	logger, err := logger.NewLogger2(logOutput, logLevel)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error initializing logger: %v", err)
	}
	// Initialize Repository
	host, portDB, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := repository.NewRepository(host, user, password, dbname, sslmode,
		timezone, portDB, timeout, schema)
	if err != nil {
		logger.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("error initializing repository: %v", err)
	}
	// Set the timezone environment variable
	os.Setenv("TZ", timezone)

	// Initialize Publisher (NATS or Noop)
	var publisher port.CustomerEventPublisher
	natsURL, natsEnabled := cfg.GetNATSData()
	if natsEnabled && natsURL != "" {
		natsPub, err := messaging.NewNATSPublisher(natsURL, logger)
		if err != nil {
			logger.IPrintf(1, "Warning: failed to connect to NATS (%v). Using NoopPublisher.", err)
			publisher = messaging.NewNoopPublisher()
		} else {
			publisher = natsPub
		}
	} else {
		logger.IPrintf(0, "NATS messaging is disabled. Using NoopPublisher.")
		publisher = messaging.NewNoopPublisher()
	}

	handler := http.NewHandler(repo, logger, publisher)
	return logger, repo, publisher, handler, cfg, nil
}
