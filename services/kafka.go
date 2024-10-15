// Package services kafka.go
package services

import (
	"context"       // Для использования контекста выполнения
	"encoding/json" // Для работы с JSON (сериализация сообщений)
	"errors"
	"github.com/segmentio/kafka-go"
	"go_microsvc/database"
	_ "go_microsvc/docs" // Сгенерированные Swagger-документы
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
		log.Printf("Ошибка сериализации сообщения: %v", err)
		return err
	}

	// Отправляем сообщение в Kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("key"), // Ключ сообщения, можно сделать уникальным
		Value: msg,           // Содержимое сообщения в формате JSON
	})

	// Проверяем на наличие ошибок при отправке сообщения
	if err != nil {
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
		return err
	}

	log.Println("Сообщение успешно отправлено в Kafka")
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

func StartKafkaConsumer(ctx context.Context, db *database.Database, brokers, topic string) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		//StartOffset: kafka.FirstOffset,
		Brokers:  []string{brokers},
		Topic:    topic,
		GroupID:  "my_consumer_group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Ошибка при закрытии reader: %v", err)
		}
	}()

	// Запускаем цикл чтения сообщений
	func() {
		for {
			select {
			case <-ctx.Done():
				// Если контекст завершен, выходим из цикла и завершаем работу консьюмера
				log.Println("Завершение работы Kafka consumer по запросу контекста")
				return
			default:
				// Чтение сообщения из Kafka
				m, err := reader.ReadMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						log.Println("Завершение работы Kafka consumer по запросу контекста")
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

				// Проверяем, было ли сообщение уже обработано
				//var existingMsg models.Message
				//if err := db.First(&existingMsg, "id = ?", msg.ID).Error; err == nil {
				//	log.Printf("Сообщение с ID %s уже обработано, пропускаем...", msg.ID)
				//	continue // Пропускаем, если сообщение уже существует
				//}

				// Сохранение сообщения в базу данных
				msg.Processed = true
				if err := db.Save(&msg).Error; err != nil {
					log.Printf("Ошибка при сохранении сообщения в базу данных: %v", err)
					// Если ошибка, не фиксируем смещение и возвращаемся к следующему сообщению
					continue
				} else {
					if err := reader.CommitMessages(ctx, m); err != nil {
						log.Printf("Ошибка при коммите смещения: %v", err)
						continue
					}
				}

			}
		}
	}()
}
