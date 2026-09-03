package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	APIPort    string

	TelegramBotToken string
	VKAppSecretKey   string
	JWTSigningKey    string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           getEnvOrDefault("DB_PORT", "5432"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           os.Getenv("DB_NAME"),
		APIPort:          getEnvOrDefault("SERVER_PORT", "8080"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		VKAppSecretKey:   os.Getenv("VK_APP_SECRET_KEY"),
		JWTSigningKey:    os.Getenv("JWT_SIGNING_KEY"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST env var not set")
	}

	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER env var not set")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD env var not set")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DB_NAME env var not set")
	}
	if cfg.APIPort == "" {
		return nil, fmt.Errorf("API_Port env var not set")
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN env var not set")
	}

	if cfg.VKAppSecretKey == "" {
		return nil, fmt.Errorf("VK_APP_SECRET_KEY env var not set")
	}
	if cfg.JWTSigningKey == "" {
		return nil, fmt.Errorf("JWT_SIGNING_KEY env var not set")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
