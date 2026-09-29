package main

import (
	"context"
	"fmt"
	"os"

	"billing/internal/adapter/driven"
	driverHttp "billing/internal/adapter/driver/http"
	"billing/internal/adapter/driver/messaging"
	"billing/internal/port"
	"billing/internal/service"
	"time"
)

// Main function to initialize the HTTP server
func main() {
	// Initialize Config
	cfg, err := driven.NewConfig("billing.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
	// Configure payment service timeout
	service.SetGlobalPaymentTimeout(time.Duration(cfg.GetPaymentTimeout()) * time.Second)
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
	// Taxer initialization
	taxer := driven.NewTaxer()
	// Pixer initialization
	pixer := driven.NewPixer(logger)
	// Issuer initialization
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

	// Initialize HTTP Handler
	os.Setenv("TZ", timezone)
	webAddr := cfg.GetWebAddr()
	handler := driverHttp.NewHandler(repo, logger, taxer, pixer, issuer)
	if err := handler.Run(webAddr); err != nil {
		logger.IPrintf(0, "Error running server: %v", err)
	}
	logger.IPrintf(0, "logger and database closed")
}
