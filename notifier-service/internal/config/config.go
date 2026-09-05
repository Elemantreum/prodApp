package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers       []string
	KafkaConsumerGroup string
	TelegramBotToken   string

	// Та же Postgres, что и у бэкенда — read-only доступ, чтобы резолвить
	// получателей (см. internal/repository.Recipients). Отдельной БД под
	// этот сервис нет и пока не нужно.
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		KafkaBrokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaConsumerGroup: getEnvOrDefault("KAFKA_CONSUMER_GROUP", "notifier-service"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             getEnvOrDefault("DB_PORT", "5432"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN env var not set")
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

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}
