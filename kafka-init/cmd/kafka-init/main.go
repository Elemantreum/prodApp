package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
)

func main() {
	brokers := strings.Split(getEnvOrDefault("KAFKA_BROKERS", "kafka:9092"), ",")

	topicNames, err := requiredTopicNames()
	if err != nil {
		log.Fatalf("kafka-init: %v", err)
	}

	topics := make([]kafka.TopicConfig, 0, len(topicNames))
	for _, name := range topicNames {
		topics = append(topics, kafka.TopicConfig{Topic: name, NumPartitions: 1, ReplicationFactor: 1})
	}

	if err := ensureTopics(brokers[0], topics); err != nil {
		log.Fatalf("kafka-init: %v", err)
	}

	log.Printf("kafka-init: ensured %d topics", len(topics))
}

func requiredTopicNames() ([]string, error) {
	raw := os.Getenv("KAFKA_TOPICS")
	if raw == "" {
		return nil, fmt.Errorf("KAFKA_TOPICS env var not set")
	}

	names := strings.Split(raw, ",")
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
		if names[i] == "" {
			return nil, fmt.Errorf("KAFKA_TOPICS contains an empty topic name")
		}
	}

	return names, nil
}

func ensureTopics(broker string, topics []kafka.TopicConfig) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("dial %s: %w", broker, err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("find controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer controllerConn.Close()

	if err := controllerConn.CreateTopics(topics...); err != nil {
		return fmt.Errorf("create topics: %w", err)
	}

	return nil
}

func getEnvOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
