package services

import (
	"context"       // Для использования контекста выполнения
	"encoding/json" // Для работы с JSON (сериализация сообщений)
	"errors"
	"github.com/segmentio/kafka-go"
	"go_microsvc/database"
	_ "go_microsvc/docs" // Импортируйте сгенерированные Swagger-документы
	"go_microsvc/models"
	"log" // Для логирования
	"time"
)

func SendMessage(brokers string, topic string, message interface{}) error {
	// Создать топик вручную
	// kafka-topics.sh --bootstrap-server localhost:9092 --create --topic messages_topic --partitions 1 --replication-factor 1
	// Проверить список топиков
	// kafka-topics.sh --bootstrap-server localhost:9092 --list

	// Создаем нового писателя Kafka напрямую
	writer := kafka.Writer{
		Addr:        kafka.TCP(brokers),  // Адреса брокеров Kafka
		Topic:       topic,               // Название топика
		Balancer:    &kafka.LeastBytes{}, // Балансировщик, используемый для выбора брокеров
		MaxAttempts: 3,                   // Максимальное количество попыток отправки
		Async:       false,               // Синхронный режим
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

func ReadMessages1(db *database.Database, brokers, topic string) error {
	// @Param groupID query string true "Идентификатор группы для Kafka Consumer"

	// Настраиваем Kafka Reader напрямую
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		MinBytes: 10e3, // Минимальное количество байт для чтения
		MaxBytes: 10e6, // Максимальное количество байт для чтения
	})
	// GroupID: groupID

	// Используем defer для закрытия reader с обработкой ошибок
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Ошибка при закрытии reader: %v", err)
		}
	}()

	for {
		// Читаем сообщение из Kafka
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Не удалось прочитать сообщение: %v", err)
			return err
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
		//if err := database.DB.Save(&msg).Error; err != nil {
		//	log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
		//	continue
		//}
		if err := db.Save(&msg).Error; err != nil {
			log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
			continue
		}
	}
	return nil
}

// ReadMessages2 читает сообщения из Kafka и обновляет статус сообщения в базе данных
func ReadMessages2(ctx context.Context, db *database.Database, brokers, topic string) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Ошибка при закрытии reader: %v", err)
		}
	}()

	for {
		// Читаем сообщение из Kafka с учетом контекста
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("Завершение работы Kafka consumer по запросу")
				return nil
			}
			log.Printf("Не удалось прочитать сообщение: %v", err)
			return err
		}

		// Декодируем сообщение в структуру модели Message
		var msg models.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Printf("Ошибка при десериализации сообщения: %v", err)
			continue
		}

		// Логируем полученное сообщение
		log.Printf("Получено сообщение: %s", msg.Content)

		// Обновляем статус сообщения в базе данных как "обработанное"
		msg.Processed = true
		if err := db.Save(&msg).Error; err != nil {
			log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
			continue
		}
	}
}

// ReadMessages читает определенное количество сообщений из Kafka и возвращает их как результат
func ReadMessages(ctx context.Context, db *database.Database, brokers, topic string, messageCount int) ([]models.Message, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Ошибка при закрытии reader: %v", err)
		}
	}()

	var messages []models.Message

	for i := 0; i < messageCount; i++ {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("Завершение работы Kafka consumer по запросу")
				return nil, nil
			}
			log.Printf("Не удалось прочитать сообщение: %v", err)
			return nil, err
		}

		var msg models.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Printf("Ошибка при десериализации сообщения: %v", err)
			continue
		}

		msg.Processed = true
		if err := db.Save(&msg).Error; err != nil {
			log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
			continue
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func StartKafkaConsumer(db *database.Database, brokers, topic string) {
	// ctx context.Context,

	reader := kafka.NewReader(kafka.ReaderConfig{
		//StartOffset: kafka.FirstOffset,
		Brokers: []string{brokers},
		Topic:   topic,
		//GroupID:     "my_consumer_group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Ошибка при закрытии reader: %v", err)
		}
	}()

	func() { // Запускаем в отдельной горутине для асинхронного выполнения
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Println("Завершение работы Kafka consumer по запросу")
					return
				}
				log.Printf("Не удалось прочитать сообщение: %v", err)
				// Задержка на 5 секунд перед следующей итерацией
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("Сообщение успешно прочитано из Kafka: Partition: %d, Offset: %d, Key: %s, Value: %s",
				m.Partition, m.Offset, string(m.Key), string(m.Value))

			var msg models.Message
			if err := json.Unmarshal(m.Value, &msg); err != nil {
				log.Printf("Ошибка при десериализации сообщения: %v", err)
				continue
			}

			msg.Processed = true
			if err := db.Save(&msg).Error; err != nil {
				log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
				continue
			}

			log.Printf("Сообщение сохранено: %s", msg.Content)

		}
	}()
}
