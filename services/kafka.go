package services

import (
	"context"       // Для использования контекста выполнения
	"encoding/json" // Для работы с JSON (сериализация сообщений)
	"github.com/segmentio/kafka-go"
	"log" // Для логирования
)

// SendMessage Функция для отправки сообщения в Kafka
func SendMessage(topic string, message interface{}) error {
	// Создаем нового писателя Kafka напрямую
	writer := kafka.Writer{
		Addr:        kafka.TCP("localhost:9092"), // Адреса брокеров Kafka
		Topic:       topic,                       // Название топика
		Balancer:    &kafka.LeastBytes{},         // Балансировщик, используемый для выбора брокеров
		MaxAttempts: 3,                           // Максимальное количество попыток отправки
		Async:       false,                       // Синхронный режим
	}
	// Используем defer для закрытия writer с обработкой ошибок
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("Ошибка при закрытии writer: %v", err)
		}
	}()

	// Преобразуем сообщение в JSON
	msg, err := json.Marshal(message)
	if err != nil {
		// Логируем ошибку сериализации и возвращаем её для дальнейшей обработки
		log.Printf("Ошибка сериализации сообщения: %v", err)
		return err // Возвращаем ошибку для обработки в вызывающей функции
	}

	// Отправляем сообщение в Kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("key"), // Ключ сообщения, можно сделать уникальным
		Value: msg,           // Содержимое сообщения в формате JSON
	})

	// Проверяем на наличие ошибок при отправке сообщения
	if err != nil {
		// Логируем ошибку отправки и возвращаем её
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
		return err
	}

	// Логируем успешную отправку
	log.Println("Сообщение успешно отправлено в Kafka")
	return nil
}
