package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Environment           string
	KafkaBootstrapServers string
	PostgresUser          string
	PostgresPassword      string
	PostgresDB            string
	PostgresHost          string
	PostgresPort          string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return Config{
		Environment:           os.Getenv("ENVIRONMENT"),
		KafkaBootstrapServers: os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		PostgresUser:          os.Getenv("POSTGRES_USER"),
		PostgresPassword:      os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:            os.Getenv("POSTGRES_DB"),
		PostgresHost:          os.Getenv("POSTGRES_HOST"),
		PostgresPort:          os.Getenv("POSTGRES_PORT"),
	}
}
