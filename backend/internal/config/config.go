package config

import (
	"os"
)

// Config holds the global application configuration.
// It contains environment-specific settings loaded from environment variables.
type Config struct {
	// Env specifies the application environment ("development" or "production")
	Env string
	// Port is the HTTP server port number
	Port string
	// DataPath is the directory path where application data is stored (database and logs)
	DataPath string
	// DLPath is the directory path where downloaded files are saved
	DLPath string
}

// Cfg is the global configuration instance, accessible throughout the application.
var Cfg *Config

// New creates and initializes a new Config instance from environment variables.
func Load() {
	Cfg = &Config{
		Env:      getEnv("APP_ENV", "development"),
		Port:     getEnv("APP_PORT", "3000"),
		DLPath:   getEnv("APP_DOWNLOAD_PATH", "./downloads"),
		DataPath: getEnv("APP_DATA_PATH", "./data"),
	}
}

// IsProd checks if the Env field is set to "production"
func (c *Config) IsProd() bool {
	return c.Env == "production"
}

// getEnv retrieves an environment variable value by key.
// If the environment variable is not set, it returns the fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
