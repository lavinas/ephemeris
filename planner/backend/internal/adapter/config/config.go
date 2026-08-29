package config

import (
	"encoding/json"
	"os"
)

// Config represents the configuration structure for the application
type Config struct {
	DB  DBConfig  `json:"db"`
	Log LogConfig `json:"log"`
	Web WebConfig `json:"web"`
}

// DBConfig represents the database configuration structure
type DBConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DBName         string `json:"dbname"`
	SSLMode        string `json:"sslmode"`
	TimeZone       string `json:"timezone"`
	ConnectTimeout int    `json:"connect_timeout"`
	BillingSchema  string `json:"billing_schema"`
}

// LogConfig represents the logging configuration structure
type LogConfig struct {
	Output string `json:"output"`
	Level  int    `json:"level"`
}

// WebConfig represents the web server configuration structure
type WebConfig struct {
	Addr string `json:"addr"`
}

// LoadConfig reads the configuration from a JSON file and unmarshals it into a Config
func NewConfig(path string) (*Config, error) {
	// Attempt to read the configuration file, if it fails, use the default configuration string
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON data into the Config struct
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetDBData returns the database configuration data as a JsonDBConfig struct
func (v *Config) GetDBData() (host string, port int, user string, password string,
	dbname string, sslmode string, timezone string, connect_timeout int, billing_schema string) {
	return v.DB.Host, v.DB.Port, v.DB.User, v.DB.Password, v.DB.DBName, v.DB.SSLMode,
		v.DB.TimeZone, v.DB.ConnectTimeout, v.DB.BillingSchema
}

// GetDBTimeZone returns the database time zone from the configuration
func (v *Config) GetDBTimeZone() string {
	return v.DB.TimeZone
}

// GetConfigData returns the entire configuration as a Config struct
func (v *Config) GetConfigData() (output string, level int) {
	return v.Log.Output, v.Log.Level
}

// GetLogOutput returns the log output from the configuration
func (v *Config) GetLogData() (output string, level int) {
	return v.Log.Output, v.Log.Level
}

// GetWebAddr returns the web server address from the configuration
func (v *Config) GetWebAddr() string {
	return v.Web.Addr
}
