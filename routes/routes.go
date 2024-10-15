// Package routes routes/routes.go
package routes

import (
	"github.com/gofiber/fiber/v2" // Импортируем Fiber
	"go_microsvc/database"        // Подключение к базе данных
	"go_microsvc/handlers"        // Импортируем пакет с обработчиками
)

// SetupRoutes инициализирует все маршруты для API
func SetupRoutes(app *fiber.App, db *database.Database) {
	api := app.Group("/api")

	api.Post("/message", func(c *fiber.Ctx) error {
		return handlers.CreateMessage(c, db) // Вызов обработчика для создания сообщения
	})

	api.Get("/stats", func(c *fiber.Ctx) error {
		return handlers.GetMessageStats(c, db) // Вызов обработчика для получения статистики
	})

	api.Get("/messages", func(c *fiber.Ctx) error {
		return handlers.GetMessages(c, db) // Вызов обработчика для получения сообщений из базы данных
	})
}
