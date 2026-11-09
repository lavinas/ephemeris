package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"pluggy/internal/config"
	"pluggy/internal/db"
	"pluggy/internal/handler"
	"pluggy/internal/logger"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Caminho do arquivo de configuração yaml")
	runMigration := flag.Bool("migrate", true, "Executar migração automática do banco")
	syncOnStartup := flag.Bool("sync", false, "Executar sincronização logo na inicialização")
	flag.Parse()

	// 1. Carregar Configuração
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Falha crítica ao carregar configurações: %v\n", err)
		os.Exit(1)
	}

	// 2. Inicializar Logger estruturado com rotação de arquivos (padrão ephemeris)
	appLogger, err := logger.NewLogger(cfg.Log.Output, cfg.Log.Level)
	if err != nil {
		fmt.Printf("Falha ao inicializar logger em %s: %v\n", cfg.Log.Output, err)
		os.Exit(1)
	}
	defer appLogger.Close()

	appLogger.Info("==================================================")
	appLogger.Info(" Iniciando Aplicação Piloto Pluggy Open Finance  ")
	appLogger.Info("==================================================")
	appLogger.Infof("[Config] Servidor configurado na porta %d", cfg.Server.Port)
	appLogger.Infof("[Config] Log gravando em: %s (com rotação diária)", cfg.Log.Output)
	appLogger.Infof("[Config] Pluggy API Base: %s (Client ID: %s)", cfg.Pluggy.BaseURL, cfg.Pluggy.ClientID)
	appLogger.Infof("[Config] Banco Postgres: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// 3. Conectar ao PostgreSQL
	database, err := db.Connect(cfg.Database.DSN())
	if err != nil {
		appLogger.Errorf("Falha ao conectar no banco PostgreSQL: %v", err)
		os.Exit(1)
	}
	defer database.Close()
	appLogger.Info("[Database] Conexão com PostgreSQL estabelecida com sucesso!")

	// 4. Executar Migrações
	if *runMigration {
		migrationPath := filepath.Join("migration", "pluggy.sql")
		if _, err := os.Stat(migrationPath); err == nil {
			appLogger.Infof("[Migration] Executando script de migração: %s", migrationPath)
			if err := database.RunMigrations(migrationPath); err != nil {
				appLogger.Warnf("[Migration] Aviso ao executar migração: %v", err)
			} else {
				appLogger.Info("[Migration] Migração validada e aplicada com sucesso.")
			}
		}
	}

	// 5. Inicializar Cliente Pluggy com Telemetria e Serviço de Sincronização
	pluggyClient := pluggy.NewClient(cfg.Pluggy.BaseURL, cfg.Pluggy.ClientID, cfg.Pluggy.ClientSecret, appLogger)
	syncer := service.NewSyncer(database, pluggyClient, appLogger)

	// Sincronização inicial opcional
	if *syncOnStartup {
		appLogger.Info("[Startup] Disparando sincronização inicial de todas as contas...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			stats, err := syncer.SyncAll(ctx)
			if err != nil {
				appLogger.Errorf("[Startup] Erro na sincronização inicial: %v", err)
			} else {
				appLogger.Infof("[Startup] Sincronização concluída: %+v", stats)
			}
		}()
	}

	// 6. Configurar Roteador HTTP e Handlers
	templatesDir := "templates"
	router, err := handler.NewRouter(cfg, database, pluggyClient, syncer, appLogger, templatesDir)
	if err != nil {
		appLogger.Errorf("Erro ao carregar roteador e templates: %v", err)
		os.Exit(1)
	}

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Iniciar Servidor HTTP em Goroutine
	go func() {
		appLogger.Infof("[Server] Servidor web HTMX ouvindo em http://localhost%s", serverAddr)
		appLogger.Infof("[Server] Endpoint Webhook ativo em http://localhost%s/api/webhook", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Errorf("Erro no servidor HTTP: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("[Server] Encerrando servidor graciosamente...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Errorf("Erro durante o encerramento do servidor: %v", err)
	}
	appLogger.Info("[Server] Servidor finalizado.")
}
