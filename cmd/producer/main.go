package main

import (
	"go_microsvc/config"   // Импортируем модуль для работы с конфигурацией
	"go_microsvc/services" // Импортируем сервис для отправки сообщений
	"log"                  // Для логирования
)

func main() {
	// Создать топик вручную
	// kafka-topics.sh --bootstrap-server localhost:9092 --create --topic messages_topic --partitions 1 --replication-factor 1
	// Проверить список топиков
	// kafka-topics.sh --bootstrap-server localhost:9092 --list

	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Создаем сообщение для отправки в Kafka
	message := map[string]string{"content": "Hello Kafka"}

	// Отправляем сообщение в Kafka, используя конфигурацию
	err := services.SendMessage(cfg.KafkaBootstrapServers, "message_topic", message)

	// Проверяем на наличие ошибки при отправке
	if err != nil {
		log.Fatalf("Ошибка отправки сообщения: %v", err)
	}

	// Логируем успешную отправку
	log.Println("Message sent to Kafka")
}
