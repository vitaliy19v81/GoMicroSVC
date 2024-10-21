// cmd/api/main.go
package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger" // Импортируем пакет для Swagger
	"github.com/swaggo/fiber-swagger"
	"go_microsvc/config"   // Импортируем пакет для загрузки конфигурации
	"go_microsvc/database" // Импортируем пакет для подключения к базе данных
	_ "go_microsvc/docs"   // Импортируйте сгенерированные Swagger-документы
	"go_microsvc/routes"
	"go_microsvc/services"
	"log" // Импортируем пакет для логирования
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.LoadConfig()

	// 2. Создание контекста с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Гарантируем, что контекст будет отменен при выходе из main

	//// Запускаем команду для создания топика
	//err := createKafkaTopic(cfg.KafkaBootstrapServers, cfg.KafkaTopic, 1, 1)
	//if err != nil {
	//	log.Fatalf("Ошибка при создании топика: %v", err)
	//}

	//if err := createTopic(cfg.KafkaBootstrapServers, cfg.KafkaTopic); err != nil {
	//	log.Fatalf("Ошибка при создании топика: %v", err)
	//}
	//log.Println("Топик успешно создан!")

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

//func createTopic(brokers string, topic string) error {
//	conn, err := kafka.Dial("tcp", brokers)
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//
//	// Здесь вы можете настроить количество партиций и реплик
//	partitions := 1
//	replicationFactor := 1
//
//	if err := conn.CreatePartitions(topic, partitions, replicationFactor); err != nil {
//		return err
//	}
//	return nil
//}

//func addKafkaToPath() {
//	// Добавляем путь к Kafka в переменную окружения PATH
//	err := os.Setenv("PATH", "/usr/local/kafka/bin:"+os.Getenv("PATH"))
//	if err != nil {
//		return
//	}
//}
//
//// createKafkaTopic создает топик в Kafka
//func createKafkaTopic(bootstrapServer, topic string, replicationFactor, partitions int) error {
//	addKafkaToPath()
//
//	// Формируем команду для создания топика
//	cmd := exec.Command("kafka-topics.sh",
//		"--create",
//		"--bootstrap-server", bootstrapServer,
//		"--replication-factor", strconv.Itoa(replicationFactor),
//		"--partitions", strconv.Itoa(partitions),
//		"--topic", topic)
//
//	// Выполняем команду и получаем стандартный вывод и ошибки
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		return err
//	}
//
//	// Логируем вывод команды
//	log.Println(string(output))
//	return nil
//}
