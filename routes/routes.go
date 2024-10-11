// Package routes routes/routes.go
package routes

import (
	"github.com/gofiber/fiber/v2" // Импортируем Fiber
	"go_microsvc/database"        // Подключение к базе данных
	"go_microsvc/handlers"        // Импортируем пакет с обработчиками
)

// SetupRoutes инициализирует все маршруты для API
func SetupRoutes(app *fiber.App, db *database.Database) {
	//func SetupRoutes(api fiber.Router, db *gorm.DB) {

	app.Post("/api/message", func(c *fiber.Ctx) error {
		return handlers.CreateMessage(c, db) // Вызов обработчика для создания сообщения
	})

	app.Get("/api/stats", func(c *fiber.Ctx) error {
		return handlers.GetMessageStats(c, db) // Вызов обработчика для получения статистики
	})

	app.Get("/api/messages", func(c *fiber.Ctx) error {
		return handlers.GetMessages(c, db) // Вызов обработчика для получения сообщений из базы данных
	})

	//app.Get("/kafka/read", func(c *fiber.Ctx) error {
	//	return handlers.GetMessagesKafka(c, db) // Вызов обработчика для получения сообщений из базы данных
	//})
	//
	//app.Get("/kafka/stop", func(c *fiber.Ctx) error {
	//	return handlers.StopKafkaConsumer(c)
	//})
}
