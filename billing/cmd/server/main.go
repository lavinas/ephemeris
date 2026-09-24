package main

import (
	"context"
	"fmt"
	"os"

	"billing/internal/adapter/driven"
	"billing/internal/adapter/driver"
	"billing/internal/adapter/driver/messaging"
	"billing/internal/port"
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
	logger, err := driven.NewLogger2(logOutput, logLevel)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}
	defer logger.Close()
	// Initialize Repository
	host, portDB, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := driven.NewPostgresRepository(host, user, password, dbname, sslmode,
		timezone, portDB, timeout, schema)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		return
	}
	defer repo.Close()
	// Issuer initialization
	taxer := driven.NewTaxer()
	// Biller initialization
	pixer := driven.NewPixer(logger)
	// issuer initialization
	issuer := driven.NewIssuer()

	// Initialize Messaging Consumer (NATS or Noop)
	var consumer port.EventConsumer
	natsURL, queueGroup, natsEnabled := cfg.GetNATSData()
	if natsEnabled && natsURL != "" {
		natsConsumer := messaging.NewNATSConsumer(natsURL, queueGroup, repo, logger)
		if err := natsConsumer.Start(context.Background()); err != nil {
			logger.IPrintf(1, "Warning: failed to start NATS consumer (%v). Using NoopConsumer.", err)
			consumer = messaging.NewNoopConsumer()
		} else {
			consumer = natsConsumer
		}
	} else {
		logger.IPrintf(0, "NATS messaging is disabled. Using NoopConsumer.")
		consumer = messaging.NewNoopConsumer()
	}
	defer func() {
		consumer.Close()
		logger.IPrintf(0, "Messaging consumer closed")
	}()

	// Initialize API Handler
	os.Setenv("TZ", timezone)
	logger.IPrintf(0, "starting API server on :8081")
	apiHandler := driver.NewAPIHandler(":8081", logger, repo, taxer, pixer, issuer)
	apiHandler.Run(":8081")
	logger.IPrintf(0, "logger and database closed")
}
