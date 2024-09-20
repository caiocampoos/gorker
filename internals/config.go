package internals

import "os"

type RabbitConfig struct {
	RABBITMQ_HOST string
	RABBITMQ_USER string
	RABBITMQ_PASS string
}

func getEnvOrDefault(name string, defValue string) string {
	value := os.Getenv(name)

	if value != "" {
		return value
	}

	return defValue
}

func RabbitMqConfigGet() RabbitConfig {
	return RabbitConfig{
		RABBITMQ_HOST: getEnvOrDefault("RABBITMQ_HOST", "127.0.0.1"),
		RABBITMQ_USER: getEnvOrDefault("RABBITMQ_USER", "guest"),
		RABBITMQ_PASS: getEnvOrDefault("RABBITMQ_PASS", "guest"),
	}
}
