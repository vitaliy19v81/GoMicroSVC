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
