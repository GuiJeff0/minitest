package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the runtime configuration for minitest-api loaded from environment variables.
type Config struct {
	AppEnv          string
	HTTPPort        string
	ShutdownTimeout time.Duration
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
}

// Load reads configuration from environment variables with sensible defaults and validates required fields.
func Load() (Config, error) {
	dbPort, err := getEnvInt("DB_PORT", 5432)
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	shutdownSeconds, err := getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SECONDS: %w", err)
	}

	cfg := Config{
		AppEnv:          getEnv("APP_ENV", "dev"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		ShutdownTimeout: time.Duration(shutdownSeconds) * time.Second,
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          dbPort,
		DBUser:          getEnv("DB_USER", "minitest"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          getEnv("DB_NAME", "minitest"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks that mandatory configuration fields are well-formed.
func (c Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT must not be empty")
	}
	if _, err := strconv.Atoi(c.HTTPPort); err != nil {
		return fmt.Errorf("HTTP_PORT must be numeric: %w", err)
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST must not be empty")
	}
	if c.DBPort <= 0 || c.DBPort > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER must not be empty")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME must not be empty")
	}
	return nil
}

// DSN builds the PostgreSQL connection string from the configuration.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
