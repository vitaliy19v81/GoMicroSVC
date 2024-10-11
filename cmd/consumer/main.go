package main

import (
	"go_microsvc/config"
	"go_microsvc/database"
	"go_microsvc/services"
	"log"
)

func main() {
	// Загрузка конфигурации из файла или переменных окружения
	cfg := config.LoadConfig()

	// Подключение к базе данных PostgreSQL с использованием параметров из конфигурации
	db, err := database.ConnectDB(cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Настраиваем Kafka Reader и передаем его в сервис для обработки сообщений
	err = services.ReadMessages1(db, cfg.KafkaBootstrapServers, cfg.KafkaTopic)
	// "my_group" группа убрана пока
	if err != nil {
		log.Fatalf("Ошибка чтения сообщений из Kafka: %v", err)
	}
}
