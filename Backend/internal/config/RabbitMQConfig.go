package config

import "os"

const defaultRabbitMQURL = "amqp://guest:guest@localhost:5672/"

type RabbitMQConfig struct {
	URL string
}

func NewRabbitMQConfig() *RabbitMQConfig {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = defaultRabbitMQURL
	}
	return &RabbitMQConfig{URL: url}
}
