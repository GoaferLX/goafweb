package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Port     int           `json:"port"`     // Port to run app on
	Env      string        `json:"env"`      // Environment i.e. production/development
	PWPepper string        `json:"pwPepper"` // For passwords
	HMACKey  string        `json:"hmacKey"`  // For hashing rememberTokens
	Database dbConfig      `json:"database"` // Database information
	Mailgun  mailgunConfig `json:"mailgun"`  // Mailgun config
	Paypal   paypalConfig  `json:"paypal"`   // Paypal integration config
}

// Config values by default if user does not provide a config file
func defaultConfig() Config {
	return Config{
		Port:     3000,                   // standard for localhost
		Env:      "dev",                  // default to development
		PWPepper: "secret-random-string", // random dev assignment
		HMACKey:  "secret-hmac-key",      // random dev assignment
		Database: defaultDBConfig(),      // Defaults to dev database
	}
}

// Loads config provided by user
// If none provided, loads default config values
// Config file must be provided if configRequired is set to true
// App will Fatal(err) if one not provided - used in prod
func LoadConfig(configRequired bool) Config {
	file, err := os.Open(".config")
	if err != nil {
		if configRequired {
			log.Fatal("Config file must be provided in production environment!")
		}
		log.Println(fmt.Errorf("Could not open config file: %w", err))
		fmt.Println("Using default config...")
		return defaultConfig()
	}
	dec := json.NewDecoder(file)
	var cfg Config
	err = dec.Decode(&cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("Could not parse config file: %w", err))
	}
	file.Close()
	fmt.Println("Using specified config")
	return cfg
}

// Returns true or false - used throughout app to protect values
func (c Config) isProd() bool {
	return c.Env == "prod"
}

type mailgunConfig struct {
	Domain       string `json:"domain"`
	APIKey       string `json:"api_key"`
	PublicAPIKey string `json:"public_api_key"`
}

type paypalConfig struct {
	ClientID string `json:"client_id"`
	Secret   string `json:"secret"`
}

// Database configuration
// Not all fields will be used dependent on database dialect being used
type dbConfig struct {
	User    string `json:"user"`
	Passwd  string `json:"passwd"`
	Net     string `json:"net"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	DBName  string `json:"dbName"`
	Dialect string `json:"dialect"`
}

// Database config to be used if one not provided by user
func defaultDBConfig() dbConfig {
	return dbConfig{
		User:    "root",
		Passwd:  "",
		Net:     "tcp",
		Host:    "localhost",
		Port:    3306,
		DBName:  "goafweb",
		Dialect: "mysql",
	}
}

// Returns a dsn string for the selected database dialect
func (dbcfg *dbConfig) dsn() string {
	switch dbcfg.Dialect {
	case "mysql":
		// user:password@tcp(host:port)/dbname
		return fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true&charset=utf8mb4,utf8", dbcfg.User, dbcfg.Passwd, dbcfg.Net, dbcfg.Host, dbcfg.Port, dbcfg.DBName)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dbcfg.Host, dbcfg.Port, dbcfg.User, dbcfg.Passwd, dbcfg.DBName)
	default:
		log.Fatal("Database drivers not supported")
		return ""
	}
}
