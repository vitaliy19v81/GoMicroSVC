package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go_microsvc/config"
	"go_microsvc/database"
	"go_microsvc/models"
	"go_microsvc/services" // Импортируем сервис для работы с Kafka
	"log"
	"net/http"
	"strconv"
	"sync"
)

var (
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
)

// CreateMessage создает новое сообщение и сохраняет его в базе данных
// @Summary Создание сообщения
// @Description Создает новое сообщение и сохраняет его в базе данных
// @Tags Api
// @Accept json
// @Produce json
// @Param message body models.CreateMessageRequest true "Сообщение"
// @Success 201 {object} models.CreateMessageResponse
// @Failure 400 {string} string "Неверный формат данных"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/message [post]
func CreateMessage(c *fiber.Ctx, db *database.Database) error {

	// Загрузка конфигурации из файла или переменных окружения
	cfg := config.LoadConfig()

	var request models.CreateMessageRequest

	// Парсинг тела запроса в структуру CreateMessageRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request: %v", err)
		return c.Status(http.StatusBadRequest).SendString("Invalid input: " + err.Error())
	}
	log.Printf("Parsed content: %s", request.Content)

	// Создание экземпляра модели сообщения
	msg := models.Message{
		Content:   request.Content,
		Processed: false, // Статус по умолчанию
	}

	// Сохраняем сообщение в базу данных
	if err := db.Create(&msg).Error; err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Ошибка сохранения сообщения: " + err.Error())
	}

	// Отправка сообщения в Kafka
	if err := services.SendMessage(cfg.KafkaBrokers, cfg.KafkaTopic, msg); err != nil {
		log.Printf("Ошибка отправки сообщения в Kafka: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Ошибка отправки сообщения в Kafka: " + err.Error())
	}

	// Возвращаем сохраненное сообщение в ответе
	return c.Status(http.StatusCreated).JSON(msg)
}

// GetMessageStats возвращает количество обработанных сообщений
// Маршрут для получения статистики обработанных сообщений
// @Summary Получение статистики
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
// @Summary Получение списка сообщений
// @Description Возвращает список сообщений с учетом offset и limit
// @Tags Api
// @Produce json
// @Param offset query int false "Смещение" default(0)
// @Param limit query int false "Лимит" default(10)
// @Success 200 {array} models.Message
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/messages [get]
func GetMessages(c *fiber.Ctx, db *database.Database) error {

	// Получение значений offset и limit из query parameters
	offsetParam := c.Query("offset", "0")
	limitParam := c.Query("limit", "10")

	// Преобразование параметров в целые числа
	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		log.Printf("Invalid offset: %v", err)
		return c.Status(http.StatusBadRequest).SendString("Invalid offset parameter")
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		log.Printf("Invalid limit: %v", err)
		return c.Status(http.StatusBadRequest).SendString("Invalid limit parameter")
	}

	// Извлечение сообщений из базы данных с использованием offset и limit
	var messages []models.Message
	if err := db.Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		log.Printf("Error retrieving messages: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Ошибка получения сообщений: " + err.Error())
	}

	// Возвращаем список сообщений в ответе
	return c.Status(http.StatusOK).JSON(messages)
}

//// GetMessagesKafka ReadMessages читает сообщения из Kafka и обновляет статус сообщения в базе данных
////
//// @Summary Чтение сообщений из Kafka
//// @Description Функция для чтения сообщений из указанного топика Kafka и обновления их статуса в базе данных
//// @Tags Kafka
//// @Param brokers query string true "Адреса брокеров Kafka"
//// @Param topic query string true "Название топика в Kafka"
//// @Success 200 {string} string "Сообщения успешно прочитаны и обновлены"
//// @Failure 400 {string} string "Ошибка чтения сообщения"
//// @Failure 500 {string} string "Ошибка сохранения сообщения в базу данных"
//// @Router /kafka/read [get]
//func GetMessagesKafka(c *fiber.Ctx, db *database.Database) error {
//	cfg := config.LoadConfig()
//
//	// Указываем количество сообщений, которое хотим прочитать из Kafka
//	messageCount := 5
//
//	messages, err := services.ReadMessages(c.Context(), db, cfg.KafkaBootstrapServers, cfg.KafkaTopic, messageCount)
//	if err != nil {
//		log.Printf("Ошибка чтения сообщений из Kafka: %v", err)
//		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
//			"error": "Ошибка при чтении сообщений из Kafka",
//		})
//	}
//
//	return c.Status(http.StatusOK).JSON(messages)
//}
//
//func GetMessagesKafka2(c *fiber.Ctx, db *database.Database) error {
//	// Загрузка конфигурации из файла или переменных окружения
//	cfg := config.LoadConfig()
//
//	// Создаем контекст для управления завершением работы Kafka consumer
//	ctx, cancel := context.WithCancel(context.Background())
//	cancelFunc = cancel // Сохраняем функцию отмены, чтобы остановить consumer при необходимости
//
//	wg.Add(1) // Увеличиваем счетчик горутин
//	go func() {
//		defer wg.Done()
//		err := services.ReadMessages2(ctx, db, cfg.KafkaBootstrapServers, cfg.KafkaTopic)
//		if err != nil {
//			log.Printf("Ошибка чтения сообщений из Kafka: %v", err)
//		}
//	}()
//
//	return c.Status(fiber.StatusOK).JSON(fiber.Map{
//		"message": "Kafka consumer запущен в фоновом режиме",
//	})
//}
//
//// StopKafkaConsumer останавливает работу Kafka consumer
//func StopKafkaConsumer(c *fiber.Ctx) error {
//	if cancelFunc != nil {
//		cancelFunc() // Останавливаем consumer
//		wg.Wait()    // Ждем завершения всех горутин
//	}
//
//	return c.Status(fiber.StatusOK).JSON(fiber.Map{
//		"message": "Kafka consumer остановлен",
//	})
//}
