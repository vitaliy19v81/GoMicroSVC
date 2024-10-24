// Package handlers handlers.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go_microsvc/config"
	"go_microsvc/database"
	"go_microsvc/models"
	"go_microsvc/services" // Импортируем сервис для работы с Kafka
	"log"
	"net/http"
	"strconv"
)

// CreateMessage создает новое сообщение и сохраняет его в базе данных
// @Summary Создание сообщения
// @Description Создает новое сообщение и сохраняет его в базе данных
// @Tags Api
// @Accept json
// @Produce json
// @Param message body models.CreateMessageRequest true "Сообщение"
// @Success 201 {object} models.CreateMessageResponse
// @Failure 400 {object} fiber.Map "Неверный формат данных"
// @Failure 422 {object} fiber.Map "Ошибка валидации данных"
// @Failure 500 {object} fiber.Map "Ошибка сервера или Kafka"
// @Router /api/message [post]
func CreateMessage(c *fiber.Ctx, db *database.Database) error {

	// Загрузка конфигурации из файла или переменных окружения
	cfg := config.LoadConfig()

	var request models.CreateMessageRequest

	// Парсинг тела запроса в структуру CreateMessageRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request: %v", err)
		// Возвращаем статус 400 и сообщение об ошибке
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input: " + err.Error()})
	}
	// Валидация поля Content
	if len(request.Content) == 0 {
		// Возвращаем статус 422 и сообщение об ошибке
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Content is required"})
	}

	log.Printf("Parsed content: %s", request.Content)

	// Создание экземпляра модели сообщения
	msg := models.Message{
		Content:   request.Content,
		Processed: false, // Статус по умолчанию
	}

	// Сохраняем сообщение в базу данных
	if err := db.Create(&msg).Error; err != nil {
		// Возвращаем статус 500 и сообщение об ошибке
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Database error: " + err.Error()})
	}

	// Отправка сообщения в Kafka
	if err := services.SendMessage(cfg.KafkaBrokers, cfg.KafkaTopic, msg); err != nil {
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
		// Возвращаем статус 500 и сообщение об ошибке
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Kafka error: " + err.Error()})
	}

	// Возвращаем сохраненное сообщение в ответе
	return c.Status(http.StatusCreated).JSON(msg)
}

// GetMessageStats возвращает количество обработанных сообщений
// Маршрут для получения статистики обработанных сообщений
// @Summary Получение статистики обработанных сообщений consumer-ом
// @Description Получает количество обработанных сообщений
// @Tags Api
// @Produce json
// @Success 200 {object} map[string]int64
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/stats [get]
func GetMessageStats(c *fiber.Ctx, db *database.Database) error {

	var count int64

	// Считаем количество обработанных сообщений
	if err := db.Model(&models.Message{}).Where("processed = ?", true).Count(&count).Error; err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Ошибка получения статистики: " + err.Error())
	}

	// Возвращаем статистику
	return c.Status(http.StatusOK).JSON(map[string]int64{"processed_messages": count})
}

// GetMessages получает сообщения из базы данных с offset и limit
// @Summary Получение списка сообщений из базы данных
// @Description Возвращает список сообщений с учетом offset и limit
// @Tags Api
// @Produce json
// @Param offset query int false "Смещение" default(0)
// @Param limit query int false "Лимит" default(10)
// @Success 200 {array} models.Message "Успешное получение сообщений"
// @Failure 400 {object} fiber.Map "Неверные параметры запроса"
// @Failure 500 {object} fiber.Map "Ошибка сервера"
// @Router /api/messages [get]
func GetMessages(c *fiber.Ctx, db *database.Database) error {

	// Получение значений offset и limit из query parameters
	offsetParam := c.Query("offset", "0")
	limitParam := c.Query("limit", "10")

	// Преобразование параметров в целые числа
	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		log.Printf("Invalid offset: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid offset parameter"})
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		log.Printf("Invalid limit: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid limit parameter"})
	}

	// Извлечение сообщений из базы данных с использованием offset и limit
	var messages []models.Message
	if err := db.Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		log.Printf("Error retrieving messages: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Database error: " + err.Error()})
	}

	// Возвращаем список сообщений в ответе
	return c.Status(http.StatusOK).JSON(messages)
}
