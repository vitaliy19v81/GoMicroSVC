// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger" // Импортируем пакет для Swagger
	"github.com/segmentio/kafka-go"
	"github.com/swaggo/fiber-swagger"
	"go_microsvc/config"   // Импортируем пакет для загрузки конфигурации
	"go_microsvc/database" // Импортируем пакет для подключения к базе данных
	_ "go_microsvc/docs"   // Импортируйте сгенерированные Swagger-документы
	"go_microsvc/routes"
	"go_microsvc/services"
	"log" // Импортируем пакет для логирования
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.LoadConfig()

	// 2. Создание контекста с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Гарантируем, что контекст будет отменен при выходе из main

	// Вызов функции создания топика
	err := createKafkaTopic("kafka:9092", "new-topic-2", 1, 1)
	if err != nil {
		log.Fatalf("Ошибка: %v\n", err)
	}

	// 3. Подключение к базе данных
	db, err := database.ConnectDB(cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// 4. Запуск Kafka consumer в отдельной горутине
	go services.StartKafkaConsumer(ctx, db, cfg.KafkaBootstrapServers, cfg.KafkaTopic)

	// 5. Запуск Kafka producer в отдельной горутине
	go func() {
		if err := services.NewKafkaProducer(ctx, cfg.KafkaBootstrapServers, cfg.KafkaTopic); err != nil {
			log.Printf("Ошибка в работе Kafka Producer: %v", err)
		}
	}()

	// 6. Создание нового Fiber приложения
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html")
	})

	// Подключаем Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // Стандартный обработчик Swagger

	// Маршрут для Swagger
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	// 7. Настройка маршрутов приложения из отдельного пакета
	routes.SetupRoutes(app, db)

	// 8. Обработка сигнала завершения для корректного завершения работы
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 9. Ожидание сигнала завершения
	go func() {
		<-c
		log.Println("Завершение работы...")
		cancel() // Отмена контекста, что приведет к завершению Kafka consumer и producer
	}()

	// 10. Горутина для запуска HTTP сервера Fiber
	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
		}
	}()

	// 11. Ожидание завершения всех процессов и корректное завершение программы
	<-ctx.Done() // Ожидание отмены контекста

	log.Println("Контекст отменен. Завершение работы...")

	// 12. Закрытие приложения Fiber с передачей контекста
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Ошибка при завершении работы Fiber: %v", err)
	}

	// Ждем немного времени, чтобы дать завершиться всем горутинам
	time.Sleep(5 * time.Second)

	log.Println("Все процессы завершены. Завершение программы.")
	os.Exit(0)
}

// Функция для создания топика в Kafka
func createKafkaTopic(brokerAddress, topic string, numPartitions, replicationFactor int) error {
	log.Printf("Подключение к брокеру Kafka: %s\n", brokerAddress)

	// Устанавливаем соединение с брокером Kafka
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Kafka: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Ошибка закрытия соединения с брокером: %v\n", err)
		}
	}()

	// Получаем информацию о контроллере кластера
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("ошибка получения контроллера Kafka: %v", err)
	}
	log.Printf("Контроллер Kafka: %s:%d\n", controller.Host, controller.Port)

	// Устанавливаем соединение с контроллером
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("ошибка подключения к контроллеру Kafka: %v", err)
	}
	defer func() {
		if err := controllerConn.Close(); err != nil {
			log.Printf("Ошибка закрытия соединения с контроллером: %v\n", err)
		}
	}()

	// Конфигурация нового топика
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}

	// Создаем топик
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("ошибка создания топика: %v", err)
	}

	log.Printf("Топик успешно создан: %s\n", topic)
	return nil
}
