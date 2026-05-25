package driven

import (
	"encoding/json"
	"os"
)

// JsonConfig represents the configuration structure for the application
type JsonConfig struct {
	DB  JsonDBConfig  `json:"db"`
	Log JsonLogConfig `json:"log"`
}

// DBConfig represents the database configuration structure
type JsonDBConfig struct {
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
type JsonLogConfig struct {
	Output string `json:"output"`
	Level  int    `json:"level"`
}

// LoadJsonConfig reads the configuration from a JSON file and unmarshals it into a JsonConfig struct
func NewConfig(path string) (*JsonConfig, error) {
	// Attempt to read the configuration file, if it fails, use the default configuration string
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON data into the Config struct
	var cfg JsonConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetDBData returns the database configuration data as a JsonDBConfig struct
func (v *JsonConfig) GetDBData() (host string, port int, user string, password string, dbname string, sslmode string, timezone string,
	connect_timeout int, billing_schema string) {
	return v.DB.Host, v.DB.Port, v.DB.User, v.DB.Password, v.DB.DBName, v.DB.SSLMode, v.DB.TimeZone, v.DB.ConnectTimeout, v.DB.BillingSchema
}

// GetDBTimeZone returns the database time zone from the configuration
func (v *JsonConfig) GetDBTimeZone() string {
	return v.DB.TimeZone
}

// GetConfigData returns the entire configuration as a JsonConfig struct
func (v *JsonConfig) GetConfigData() (output string, level int) {
	return v.Log.Output, v.Log.Level
}
