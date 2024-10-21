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

//// CreateTopic создает топик в Kafka, если он еще не существует
//func CreateTopic(broker, topic string) error {
//	// Для установки библиотеки
//	// GOPROXY=https://goproxy.cn go get -u github.com/confluentinc/confluent-kafka-go/kafka
//	// И добавить алиас в импорты confluent "github.com/confluentinc/confluent-kafka-go/kafka"
//
//	// Создаем конфигурацию для администраторского клиента Kafka
//	config := &confluent.ConfigMap{
//		"bootstrap.servers": broker,
//	}
//
//	// Создаем администраторский клиент Kafka
//	adminClient, err := confluent.NewAdminClient(config)
//	if err != nil {
//		log.Printf("Ошибка при создании администраторского клиента Kafka: %v", err)
//		return err
//	}
//	defer adminClient.Close()
//
//	// Определяем параметры для создания топика
//	topicSpecifications := []confluent.TopicSpecification{
//		{
//			Topic:             topic,
//			NumPartitions:     1, // Количество партиций
//			ReplicationFactor: 1, // Коэффициент репликации
//		},
//	}
//
//	// Создаем топик с использованием администраторского клиента
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	results, err := adminClient.CreateTopics(ctx, topicSpecifications)
//	if err != nil {
//		log.Printf("Ошибка при создании топика %s: %v", topic, err)
//		return err
//	}
//
//	// Проверяем результаты создания топика
//	for _, result := range results {
//		if result.Error.Code() != confluent.ErrNoError {
//			log.Printf("Не удалось создать топик %s: %v", result.Topic, result.Error)
//		} else {
//			log.Printf("Топик %s успешно создан", result.Topic)
//		}
//	}
//
//	return nil
//}

// KafkaProducer представляет собой структуру для работы с Kafka producer
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer инициализирует новый Kafka producer и управляет его жизненным циклом
func NewKafkaProducer(ctx context.Context, brokers string, topic string) error {
	producer := &KafkaProducer{
		writer: &kafka.Writer{
			Addr:        kafka.TCP(brokers),
			Topic:       topic,
			Balancer:    &kafka.LeastBytes{},
			MaxAttempts: 3,
			Async:       true, // Асинхронный режим для повышения производительности
		},
	}

	// Используем defer для корректного завершения работы writer при завершении контекста
	defer func() {
		<-ctx.Done()
		if err := producer.writer.Close(); err != nil {
			log.Printf("Ошибка при закрытии Kafka producer: %v", err)
		}
		log.Println("Kafka producer завершил работу.")
	}()

	// Запуск основного цикла работы producer в горутине
	go func() {
		<-ctx.Done()
		log.Println("Завершение работы Kafka producer по запросу контекста")
	}()

	// Пока контекст не завершен, производим работу
	<-ctx.Done() // Блокируем до завершения контекста
	return nil
}

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

//// NewKafkaProducer создает новый Kafka Producer и запускает его
//func NewKafkaProducer(ctx context.Context, brokers []string, topic string) error {
//	writer := &kafka.Writer{
//		Addr:         kafka.TCP(brokers...), // Используем функцию TCP для адресов брокеров
//		Topic:        topic,
//		Balancer:     &kafka.LeastBytes{},   // Балансировщик для распределения сообщений
//		MaxAttempts:  3,                     // Количество попыток отправки сообщения
//		Async:        true,                  // Асинхронный режим для повышения производительности
//		BatchTimeout: 10 * time.Millisecond, // Таймаут для объединения сообщений в пакет
//	}
//
//	// Здесь добавьте вашу логику по отправке сообщений
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				log.Println("Закрытие Kafka Producer...")
//				if err := writer.Close(); err != nil {
//					log.Printf("Ошибка при закрытии Kafka Producer: %v", err)
//				}
//				return
//			default:
//				// Логика отправки сообщений
//				// Пример: writer.WriteMessages(ctx, kafka.Message{Value: []byte("message")})
//				time.Sleep(1 * time.Second) // Просто задержка для примера
//			}
//		}
//	}()
//
//	return nil
//}

//type KafkaProducer struct {
//	writer *kafka.Writer
//	mu     sync.Mutex // Для обеспечения потокобезопасности
//}
//
//// NewKafkaProducer создает новый экземпляр KafkaProducer
//func NewKafkaProducer(brokers string, topic string) *KafkaProducer {
//	return &KafkaProducer{
//		writer: &kafka.Writer{
//			Addr:        kafka.TCP(brokers),
//			Topic:       topic,
//			Balancer:    &kafka.LeastBytes{},
//			MaxAttempts: 3,
//			Async:       true, // Асинхронный режим
//		},
//	}
//}
//
//// SendMessage отправляет сообщение в Kafka
//func (p *KafkaProducer) SendMessage(message interface{}) error {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//
//	// Преобразование сообщения в байты (например, если оно в формате JSON)
//	msg, err := json.Marshal(message)
//	if err != nil {
//		log.Printf("Ошибка сериализации сообщения: %v", err)
//		return err
//	}
//
//	// Отправляем сообщение
//	return p.writer.WriteMessages(context.Background(), kafka.Message{
//		Key:   []byte("some_key"), // Опционально, вы можете использовать ключ
//		Value: msg,
//	})
//}
//
//// Close закрывает соединение с Kafka
//func (p *KafkaProducer) Close() {
//	if err := p.writer.Close(); err != nil {
//		log.Println("Ошибка закрытия Kafka writer:", err)
//	}
//}
