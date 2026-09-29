package driven

import (
	"encoding/json"
	"fmt"
	"os"
)

type JsonConfig struct {
	DB      JsonDBConfig      `json:"db"`
	Log     JsonLogConfig     `json:"log"`
	Web     JsonWebConfig     `json:"web"`
	NATS     JsonNATSConfig     `json:"nats"`
	Service  JsonServiceConfig  `json:"service"`
	AutoBill JsonAutoBillConfig `json:"auto_bill"`
}

// JsonAutoBillConfig represents the automatic billing cron configuration
type JsonAutoBillConfig struct {
	Enabled       *bool    `json:"enabled"`
	DaysInAdvance int      `json:"days_in_advance"`
	Hours         []int    `json:"hours"`
	Schedules     []string `json:"schedules"`
}

// JsonServiceConfig represents the service layer configuration structure
type JsonServiceConfig struct {
	PaymentTimeout int `json:"payment_timeout"`
}

// JsonDBConfig represents the database configuration structure
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

// JsonLogConfig represents the logging configuration structure
type JsonLogConfig struct {
	Output string `json:"output"`
	Level  int    `json:"level"`
}

// JsonWebConfig represents the web server configuration structure
type JsonWebConfig struct {
	Addr string `json:"addr"`
}

// JsonNATSConfig represents the NATS messaging configuration structure
type JsonNATSConfig struct {
	URL        string `json:"url"`
	QueueGroup string `json:"queue_group"`
	Enabled    bool   `json:"enabled"`
}

// LoadJsonConfig reads the configuration from a JSON file and unmarshals it into a JsonConfig
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
func (v *JsonConfig) GetDBData() (host string, port int, user string, password string,
	dbname string, sslmode string, timezone string, connect_timeout int, billing_schema string) {
	return v.DB.Host, v.DB.Port, v.DB.User, v.DB.Password, v.DB.DBName, v.DB.SSLMode,
		v.DB.TimeZone, v.DB.ConnectTimeout, v.DB.BillingSchema
}

// GetDBTimeZone returns the database time zone from the configuration
func (v *JsonConfig) GetDBTimeZone() string {
	return v.DB.TimeZone
}

// GetConfigData returns the entire configuration as a JsonConfig struct
func (v *JsonConfig) GetConfigData() (output string, level int) {
	return v.Log.Output, v.Log.Level
}

// GetLogOutput returns the log output from the configuration
func (v *JsonConfig) GetLogData() (output string, level int) {
	return v.Log.Output, v.Log.Level
}

// GetWebAddr returns the web server address from the configuration
func (v *JsonConfig) GetWebAddr() string {
	if v.Web.Addr == "" {
		return ":8081"
	}
	return v.Web.Addr
}

// GetNATSData returns the NATS configuration
func (v *JsonConfig) GetNATSData() (url, queueGroup string, enabled bool) {
	return v.NATS.URL, v.NATS.QueueGroup, v.NATS.Enabled
}

// GetPaymentTimeout returns the payment service timeout in seconds
func (v *JsonConfig) GetPaymentTimeout() int {
	if v.Service.PaymentTimeout <= 0 {
		return 20
	}
	return v.Service.PaymentTimeout
}

// GetAutoBillData returns the auto-bill configuration
func (v *JsonConfig) GetAutoBillData() (enabled bool, daysInAdvance int, schedules []string) {
	enabled = true
	if v.AutoBill.Enabled != nil {
		enabled = *v.AutoBill.Enabled
	}

	daysInAdvance = v.AutoBill.DaysInAdvance
	if daysInAdvance <= 0 {
		daysInAdvance = 3
	}

	// If explicit schedules are provided, use them
	if len(v.AutoBill.Schedules) > 0 {
		schedules = append(schedules, v.AutoBill.Schedules...)
	}

	// If hours are specified (e.g. [8, 14]), construct standard cron schedules ("0 H * * *")
	for _, h := range v.AutoBill.Hours {
		if h >= 0 && h <= 23 {
			schedules = append(schedules, fmt.Sprintf("0 %d * * *", h))
		}
	}

	// Default to 8:00 AM daily if no schedules or hours are configured
	if len(schedules) == 0 {
		schedules = []string{"0 8 * * *"}
	}

	return enabled, daysInAdvance, schedules
}


