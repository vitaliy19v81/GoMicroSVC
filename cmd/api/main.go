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

	// 3. Подключение к базе данных
	db, err := database.ConnectDB(cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// 4. Создание нового Fiber приложения
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html")
	})

	// Подключаем Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // Стандартный обработчик Swagger

	// Маршрут для Swagger
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	// 5. Настройка маршрутов приложения из отдельного пакета
	routes.SetupRoutes(app, db)

	// 6. Обработка сигнала завершения для корректного завершения работы
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 7. Ожидание сигнала завершения
	go func() {
		<-c
		log.Println("Завершение работы...")
		cancel() // Отмена контекста, что приведет к завершению Kafka consumer
	}()

	// 8. Запуск Kafka consumer в отдельной горутине
	go services.StartKafkaConsumer(ctx, db, cfg.KafkaBootstrapServers, cfg.KafkaTopic)

	// 9. Горутина для запуска HTTP сервера Fiber
	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
		}
	}()

	// 10. Ожидание завершения всех процессов и корректное завершение программы
	<-ctx.Done() // Ожидание отмены контекста

	log.Println("Контекст отменен. Завершение работы...")

	// 11. Закрытие приложения Fiber с передачей контекста
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Ошибка при завершении работы Fiber: %v", err)
	}

	// Ждем немного времени, чтобы дать завершиться всем горутинам
	time.Sleep(5 * time.Second)

	log.Println("Все процессы завершены. Завершение программы.")
	os.Exit(0)
}

//// cmd/api/main.go
//package main
//
//import (
//	_ "github.com/arsmn/fiber-swagger/v2" // fiber-swagger middleware
//	"github.com/gofiber/fiber/v2"
//	"github.com/gofiber/swagger" // Импортируем пакет для Swagger
//	"github.com/swaggo/fiber-swagger"
//	"go_microsvc/config"   // Импортируем пакет для загрузки конфигурации
//	"go_microsvc/database" // Импортируем пакет для подключения к базе данных
//	_ "go_microsvc/docs"   // Импортируйте сгенерированные Swagger-документы
//	"go_microsvc/models"   // Модель сообщения
//	"go_microsvc/services" // Для отправки сообщений в Kafka
//	"log"                  // Импортируем пакет для логирования
//	"net/http"             // Для работы с HTTP кодами статусов
//)
//
//// @title Example Fiber Swagger API
//// @version 1.0
//// @description This is a sample server for Swagger.
//// @termsOfService http://swagger.io/terms/
//// @contact.name API Support
//// @contact.url http://www.swagger.io/support
//// @contact.email support@swagger.io
//// @license.name Apache 2.0
//// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//// @host localhost:8080
//// @BasePath /api
//// @openapi: 3.0.0
//
//func main() {
//	// Загрузка конфигурации из файла или переменных окружения
//	cfg := config.LoadConfig()
//
//	// Подключение к базе данных PostgreSQL с использованием параметров из конфигурации
//	db, err := database.ConnectDB(cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresHost, cfg.PostgresPort)
//	if err != nil {
//		log.Fatalf("Ошибка подключения к базе данных: %v", err)
//	}
//
//	// Создаем новый Fiber апп
//	app := fiber.New()
//
//	app.Get("/", func(c *fiber.Ctx) error {
//		return c.Redirect("/swagger/index.html")
//	})
//
//	// Подключаем Swagger
//	app.Get("/swagger/*", swagger.HandlerDefault) // Стандартный обработчик Swagger
//
//	// Определение маршрутов
//	api := app.Group("/api")
//
//	// Маршрут для создания сообщений
//	// @Summary Создание сообщения
//	// @Description Создание нового сообщения и отправка его в Kafka
//	// @Tags Messages
//	// @Accept json
//	// @Produce json
//	// @Param message body models.CreateMessageRequest true "Сообщение для отправки"
//	// @Success 201 {object} models.Message "Сообщение успешно создано"
//	// @Failure 400 {string} string "Неверный формат данных"
//	// @Failure 500 {string} string "Ошибка сохранения сообщения"
//	// @Router /api/messages [post]
//	api.Post("/messages", func(c *fiber.Ctx) error {
//		var request models.CreateMessageRequest
//
//		// Парсинг тела запроса в структуру CreateMessageRequest
//		if err := c.BodyParser(&request); err != nil {
//			return c.Status(http.StatusBadRequest).SendString("Invalid input: " + err.Error())
//		}
//
//		// Создание экземпляра модели сообщения
//		msg := models.Message{
//			Content:   request.Content,
//			Processed: false, // Статус по умолчанию
//		}
//
//		// Сохраняем сообщение в базу данных
//		if err := db.Create(&msg).Error; err != nil {
//			return c.Status(http.StatusInternalServerError).SendString("Ошибка сохранения сообщения: " + err.Error())
//		}
//
//		// Отправляем сообщение в Kafka
//		err := services.SendMessage(cfg.KafkaBootstrapServers, "message_topic", msg.Content)
//		if err != nil {
//			return c.Status(http.StatusInternalServerError).SendString("Ошибка отправки сообщения в Kafka: " + err.Error())
//		}
//
//		// Помечаем сообщение как обработанное
//		msg.Processed = true
//		if err := db.Save(&msg).Error; err != nil {
//			return c.Status(http.StatusInternalServerError).SendString("Ошибка обновления статуса сообщения: " + err.Error())
//		}
//
//		// Возвращаем сохраненное сообщение в ответе
//		return c.Status(http.StatusCreated).JSON(msg)
//	})
//
//	// Маршрут для получения статистики обработанных сообщений
//	// @Summary Получение статистики обработанных сообщений
//	// @Description Получает количество обработанных сообщений
//	// @Tags Statistics
//	// @Produce json
//	// @Success 200 {object} map[string]int64 "Количество обработанных сообщений"
//	// @Failure 500 {string} string "Ошибка получения статистики"
//	// @Router /api/stats [get]
//	api.Get("/stats", func(c *fiber.Ctx) error {
//		var count int64
//
//		// Считаем количество обработанных сообщений
//		if err := db.Model(&models.Message{}).Where("processed = ?", true).Count(&count).Error; err != nil {
//			return c.Status(http.StatusInternalServerError).SendString("Ошибка получения статистики: " + err.Error())
//		}
//
//		// Возвращаем статистику
//		return c.Status(http.StatusOK).JSON(map[string]int64{"processed_messages": count})
//	})
//
//	// Маршрут для Swagger
//	api.Get("/docs/*", fiberSwagger.WrapHandler)
//
//	// Запуск HTTP сервера на порту 8080
//	if err := app.Listen(":8080"); err != nil {
//		log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
//	}
//}
