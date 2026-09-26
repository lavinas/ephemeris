package main

import (
	"context"
	"fmt"
	"os"

	"planner/internal/adapter/config"
	"planner/internal/adapter/http"
	"planner/internal/adapter/logger"
	"planner/internal/adapter/messaging"
	"planner/internal/adapter/repository"
	"planner/internal/port"
)

const (
	configFile = "planner.json"
)

// Main function to initialize the API server
func main() {
	// Initialize Config
	log, repo, handler, cfg, err := startAll()
	if err != nil {
		fmt.Printf("Error initializing components: %v\n", err)
		return
	}
	defer func() {
		log.IPrintf(0, "Logger closed")
		log.Close()
	}()
	defer func() {
		repo.Close()
		log.IPrintf(0, "Repository closed")
	}()

	// Initialize Messaging Consumer (NATS or Noop)
	var consumer port.EventConsumer
	natsURL, queueGroup, natsEnabled := cfg.GetNATSData()
	if natsEnabled && natsURL != "" {
		natsConsumer := messaging.NewNATSConsumer(natsURL, queueGroup, repo, log)
		if err := natsConsumer.Start(context.Background()); err != nil {
			log.IPrintf(1, "Warning: failed to start NATS consumer (%v). Using NoopConsumer.", err)
			consumer = messaging.NewNoopConsumer()
		} else {
			consumer = natsConsumer
		}
	} else {
		log.IPrintf(0, "NATS messaging is disabled. Using NoopConsumer.")
		consumer = messaging.NewNoopConsumer()
	}
	defer func() {
		consumer.Close()
		log.IPrintf(0, "Messaging consumer closed")
	}()

	if err := handler.Run(cfg.GetWebAddr()); err != nil {
		log.IPrintf(0, "Error running server: %v", err)
	}

}

// initComponents initializes the logger and repository components
func startAll() (*logger.Logger2, *repository.Repository, *http.Handler, *config.Config, error) {
	// Initialize Config
	cfg, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return nil, nil, nil, nil, fmt.Errorf("error loading config: %v", err)
	}
	// Initialize the logger
	logOutput, logLevel := cfg.GetLogData()
	logger, err := logger.NewLogger2(logOutput, logLevel)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error initializing logger: %v", err)
	}
	// Initialize Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := repository.NewRepository(host, user, password, dbname, sslmode,
		timezone, port, timeout, schema)
	if err != nil {
		logger.Close()
		return nil, nil, nil, nil, fmt.Errorf("error initializing repository: %v", err)
	}
	// Set the timezone environment variable
	os.Setenv("TZ", timezone)
	return logger, repo, http.NewHandler(repo, logger), cfg, nil
}
