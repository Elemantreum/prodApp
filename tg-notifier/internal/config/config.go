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
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		KafkaBrokers:       strings.Split(getEnvOrDefault("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaConsumerGroup: getEnvOrDefault("KAFKA_CONSUMER_GROUP", "tg-notifier"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN env var not set")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}
