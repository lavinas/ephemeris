package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Pluggy   PluggyConfig   `yaml:"pluggy"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type PluggyConfig struct {
	BaseURL      string `yaml:"base_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	WebhookURL   string `yaml:"webhook_url"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8085,
		},
		Pluggy: PluggyConfig{
			BaseURL: "https://api.pluggy.ai",
		},
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     5435,
			User:     "root",
			Password: "root",
			DBName:   "pluggy",
			SSLMode:  "disable",
		},
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("erro ao decodificar config.yaml: %w", err)
		}
	}

	// Environment variable overrides
	if portStr := os.Getenv("SERVER_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = p
		}
	}
	if cid := os.Getenv("PLUGGY_CLIENT_ID"); cid != "" {
		cfg.Pluggy.ClientID = cid
	}
	if csec := os.Getenv("PLUGGY_CLIENT_SECRET"); csec != "" {
		cfg.Pluggy.ClientSecret = csec
	}
	if base := os.Getenv("PLUGGY_BASE_URL"); base != "" {
		cfg.Pluggy.BaseURL = base
	}
	if wh := os.Getenv("WEBHOOK_PUBLIC_URL"); wh != "" {
		cfg.Pluggy.WebhookURL = wh
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Database.Port = p
		}
	}
	if u := os.Getenv("DB_USER"); u != "" {
		cfg.Database.User = u
	}
	if p := os.Getenv("DB_PASSWORD"); p != "" {
		cfg.Database.Password = p
	}
	if db := os.Getenv("DB_NAME"); db != "" {
		cfg.Database.DBName = db
	}
	if ssl := os.Getenv("DB_SSLMODE"); ssl != "" {
		cfg.Database.SSLMode = ssl
	}

	return cfg, nil
}

