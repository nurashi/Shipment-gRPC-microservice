package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCPort string
	DB       DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort: getEnv("GRPC_PORT", "50051"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "shipment"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "shipment_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if cfg.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if cfg.DB.Host == "" {
		return nil, fmt.Errorf("DB_HOST environment variable is required")
	}

	return cfg, nil
}

func (c *Config) GetPostgresConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
