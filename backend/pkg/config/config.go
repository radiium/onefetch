package config

import (
	"os"
)

type Config struct {
	Env      string
	Port     string
	DataPath string
	DLPath   string
}

func New() *Config {
	config := &Config{
		Env:      getEnv("APP_ENV", "development"),
		Port:     getEnv("APP_PORT", "3000"),
		DLPath:   getEnv("APP_DOWNLOAD_PATH", "./downloads"),
		DataPath: getEnv("APP_DATA_PATH", "./data"),
	}

	return config
}

func (c *Config) IsProd() bool {
	return c.Env == "production"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
