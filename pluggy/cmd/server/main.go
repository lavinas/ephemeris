package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"pluggy/internal/config"
	"pluggy/internal/db"
	"pluggy/internal/handler"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Caminho do arquivo de configuração yaml")
	runMigration := flag.Bool("migrate", true, "Executar migração automática do banco")
	syncOnStartup := flag.Bool("sync", false, "Executar sincronização logo na inicialização")
	flag.Parse()

	log.Println("==================================================")
	log.Println(" Iniciando Aplicação Piloto Pluggy Open Finance  ")
	log.Println("==================================================")

	// 1. Carregar Configuração
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Falha crítica ao carregar configurações: %v", err)
	}
	log.Printf("[Config] Servidor configurado na porta %d", cfg.Server.Port)
	log.Printf("[Config] Pluggy API Base: %s (Client ID: %s)", cfg.Pluggy.BaseURL, cfg.Pluggy.ClientID)
	log.Printf("[Config] Banco Postgres: %s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// 2. Conectar ao PostgreSQL
	database, err := db.Connect(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Falha ao conectar no banco PostgreSQL: %v", err)
	}
	defer database.Close()
	log.Println("[Database] Conexão com PostgreSQL estabelecida com sucesso!")

	// 3. Executar Migrações
	if *runMigration {
		migrationPath := filepath.Join("migration", "pluggy.sql")
		if _, err := os.Stat(migrationPath); err == nil {
			log.Printf("[Migration] Executando script de migração: %s", migrationPath)
			if err := database.RunMigrations(migrationPath); err != nil {
				log.Printf("[Migration] Aviso ao executar migração: %v", err)
			} else {
				log.Println("[Migration] Migração validada e aplicada com sucesso.")
			}
		}
	}

	// 4. Inicializar Cliente Pluggy e Serviço de Sincronização
	pluggyClient := pluggy.NewClient(cfg.Pluggy.BaseURL, cfg.Pluggy.ClientID, cfg.Pluggy.ClientSecret)
	syncer := service.NewSyncer(database, pluggyClient)

	// Sincronização inicial opcional
	if *syncOnStartup {
		log.Println("[Startup] Disparando sincronização inicial de todas as contas...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			stats, err := syncer.SyncAll(ctx)
			if err != nil {
				log.Printf("[Startup] Erro na sincronização inicial: %v", err)
			} else {
				log.Printf("[Startup] Sincronização concluída: %+v", stats)
			}
		}()
	}

	// 5. Configurar Roteador HTTP e Handlers
	templatesDir := "templates"
	router, err := handler.NewRouter(cfg, database, pluggyClient, syncer, templatesDir)
	if err != nil {
		log.Fatalf("Erro ao carregar roteador e templates: %v", err)
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
		log.Printf("[Server] Servidor web HTMX ouvindo em http://localhost%s", serverAddr)
		log.Printf("[Server] Endpoint Webhook ativo em http://localhost%s/api/webhook", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro no servidor HTTP: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Server] Encerrando servidor graciosamente...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erro durante o encerramento do servidor: %v", err)
	}
	log.Println("[Server] Servidor finalizado.")
}

