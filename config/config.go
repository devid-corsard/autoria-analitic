package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config holds app configuration from .env.
type Config struct {
	APIKey     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// DSN returns the Postgres connection string.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// LoadConfig loads .env and returns Config. Panics if required vars are missing.
func LoadConfig(log *zap.Logger) Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("loading .env failed", zap.Error(err))
	}
	cfg := Config{
		APIKey:     os.Getenv("api_key"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}
	if cfg.APIKey == "" {
		panic("api_key is required")
	}
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		panic("DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME are required")
	}
	return cfg
}
