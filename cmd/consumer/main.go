package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go_microsvc/database"
	"go_microsvc/models"
	"log"
)

func main() {
	// Настраиваем Kafka Reader напрямую с использованием именованных полей
	r := kafka.Reader{
		Addr:    kafka.TCP("localhost:9092"), // Адреса брокеров Kafka
		Topic:   "message_topic",             // Название топика
		GroupID: "my_group",                  // Идентификатор группы консьюмеров
	}

	// Используем defer для закрытия reader с обработкой ошибок
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("Ошибка при закрытии reader: %v", err)
		}
	}()

	for {
		// Читаем сообщение из Kafka
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Не удалось прочитать сообщение: %v", err)
		}

		// Декодируем сообщение из формата JSON в структуру модели Message
		var msg models.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Printf("Ошибка при десериализации сообщения: %v", err)
			continue
		}

		// Логируем полученное сообщение
		log.Printf("Получено сообщение: %s", msg.Content)

		// Обновляем статус сообщения в базе данных как "обработанное"
		msg.Processed = true
		if err := database.DB.Save(&msg).Error; err != nil {
			log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
			continue
		}
	}
}
