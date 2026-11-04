package main

import (
	"fmt"
	"os"

	"accounting/internal/adapter/bank/santander"
	"accounting/internal/adapter/config"
	"accounting/internal/adapter/cron"
	"accounting/internal/adapter/http"
	"accounting/internal/adapter/logger"
	"accounting/internal/adapter/repository"
	"accounting/internal/service"
)

const configFile = "accounting.json"

func main() {
	configPath := configFile
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config %s: %v\n", configPath, err)
		return
	}

	// 1. Logger
	logOutput, logLevel := cfg.GetLogData()
	log, err := logger.NewLogger(logOutput, logLevel)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}
	defer func() {
		log.IPrintf(0, "Accounting logger closing")
		log.Close()
	}()

	// 2. Repository
	host, port, user, password, dbname, sslmode, timezone, timeout, schema := cfg.GetDBData()
	repo, err := repository.NewRepository(host, user, password, dbname, sslmode, timezone, port, timeout, schema)
	if err != nil {
		log.IPrintf(2, "Error initializing database repository: %v", err)
		return
	}
	defer func() {
		repo.Close()
		log.IPrintf(0, "Accounting repository closed")
	}()

	os.Setenv("TZ", timezone)

	// 3. Bank Providers (Santander API, and future providers like OFX)
	santanderClient, err := santander.NewClient(cfg.Santander, log)
	if err != nil {
		log.IPrintf(1, "Warning: Santander client initialization encountered issue: %v", err)
	}

	// 4. Services
	statementListService := service.NewStatementListService(repo, log)
	statementSyncService := service.NewStatementSyncService(repo, santanderClient, log, cfg.SyncJob.DaysBack)

	// 5. Scheduled Cron Job for periodic banking statements sync
	scheduler := cron.NewScheduler(cfg.SyncJob, statementSyncService, log)
	if err := scheduler.Start(); err != nil {
		log.IPrintf(1, "Warning: failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// 6. HTTP Server (API & HTMX)
	handler := http.NewHandler(repo, log, statementListService, statementSyncService)
	if err := handler.Run(cfg.GetWebAddr()); err != nil {
		log.IPrintf(2, "Error running accounting server: %v", err)
	}
}
