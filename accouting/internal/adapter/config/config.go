package config

import (
	"encoding/json"
	"os"
	"strings"
)

// Config represents the application configuration
type Config struct {
	DB        DBConfig        `json:"db"`
	Log       LogConfig       `json:"log"`
	Web       WebConfig       `json:"web"`
	Santander SantanderConfig `json:"santander"`
	SyncJob   SyncJobConfig   `json:"sync_job"`
}

type DBConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DBName         string `json:"dbname"`
	SSLMode        string `json:"sslmode"`
	Timezone       string `json:"timezone"`
	ConnectTimeout int    `json:"connect_timeout"`
	Schema         string `json:"schema"`
}

type LogConfig struct {
	Output string `json:"output"`
	Level  int    `json:"level"`
}

type WebConfig struct {
	Addr string `json:"addr"`
}

type SantanderConfig struct {
	Environment     string `json:"environment"`      // "sandbox" or "production"
	SandboxBaseURL  string `json:"sandbox_base_url"` // Default: https://trust-sandbox.api.santander.com.br
	ProdBaseURL     string `json:"prod_base_url"`    // Default: https://trust-open.api.santander.com.br
	TokenURL        string `json:"token_url"`        // OAuth token endpoint
	BankID          string `json:"bank_id"`          // e.g. "90400888000142"
	CertificatePath string `json:"certificate_path"` // Path to .pfx or client cert
	CertPassword    string `json:"certificate_password"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
}

type SyncJobConfig struct {
	Enabled  bool   `json:"enabled"`
	CronExpr string `json:"cron_expr"` // e.g. "0 */6 * * *"
	DaysBack int    `json:"days_back"` // Default days to look back
}

// NewConfig loads configuration from path
func NewConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.setDefaults()
	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.DB.Schema == "" {
		c.DB.Schema = "accounting"
	}
	if c.DB.Timezone == "" {
		c.DB.Timezone = "America/Sao_Paulo"
	}
	if c.DB.ConnectTimeout == 0 {
		c.DB.ConnectTimeout = 10
	}
	if c.Santander.SandboxBaseURL == "" {
		c.Santander.SandboxBaseURL = "https://trust-sandbox.api.santander.com.br"
	}
	if c.Santander.ProdBaseURL == "" {
		c.Santander.ProdBaseURL = "https://trust-open.api.santander.com.br"
	}
	if c.Santander.TokenURL == "" {
		if strings.EqualFold(c.Santander.Environment, "production") || strings.EqualFold(c.Santander.Environment, "prod") {
			c.Santander.TokenURL = c.Santander.ProdBaseURL + "/auth/oauth/v2/token"
		} else {
			c.Santander.TokenURL = c.Santander.SandboxBaseURL + "/auth/oauth/v2/token"
		}
	}
	if c.Santander.BankID == "" {
		c.Santander.BankID = "90400888000142"
	}
	if c.Santander.TimeoutSeconds == 0 {
		c.Santander.TimeoutSeconds = 30
	}
	if c.SyncJob.DaysBack == 0 {
		c.SyncJob.DaysBack = 15
	}
	if c.SyncJob.CronExpr == "" {
		c.SyncJob.CronExpr = "0 */6 * * *"
	}
}

func (c *Config) GetDBData() (host string, port int, user, password, dbname, sslmode, timezone string, connectTimeout int, schema string) {
	return c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode, c.DB.Timezone, c.DB.ConnectTimeout, c.DB.Schema
}

func (c *Config) GetLogData() (output string, level int) {
	return c.Log.Output, c.Log.Level
}

func (c *Config) GetWebAddr() string {
	return c.Web.Addr
}

