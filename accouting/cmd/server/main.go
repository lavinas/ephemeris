package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	DB struct {
		Host           string `json:"host"`
		Port           int    `json:"port"`
		User           string `json:"user"`
		Password       string `json:"password"`
		DBName         string `json:"dbname"`
		SSLMode        string `json:"sslmode"`
		Timezone       string `json:"timezone"`
		ConnectTimeout int    `json:"connect_timeout"`
		Schema         string `json:"schema"`
	} `json:"db"`
	Log struct {
		Output string `json:"output"`
		Level  int    `json:"level"`
	} `json:"log"`
	Web struct {
		Addr string `json:"addr"`
	} `json:"web"`
}

func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	configPath := "accounting.json"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Printf("Aviso: Falha ao carregar %s (%v). Usando portas padrão.", configPath, err)
		cfg = &Config{}
		cfg.Web.Addr = ":8084"
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"accounting"}`))
	})

	// Static files handling if directory exists
	if _, err := os.Stat("web/static"); err == nil {
		fs := http.FileServer(http.Dir("web/static"))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	// Home endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>Accounting Service</h1><p>Sistema contábil e banking ativo.</p>")
	})

	server := &http.Server{
		Addr:         cfg.Web.Addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Iniciando accounting_api em %s...", cfg.Web.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro no servidor HTTP: %v", err)
	}
}
