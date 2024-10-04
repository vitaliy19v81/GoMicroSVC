package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5" // Используем chi для маршрутизации
	"go_microsvc/database"     // Подключение к базе данных
	"go_microsvc/models"       // Модель сообщения
	"go_microsvc/services"     // Для отправки сообщений в Kafka
	"net/http"                 // Для работы с HTTP-запросами и ответами
)

// Router создает маршруты для API
func Router() chi.Router {
	r := chi.NewRouter()

	// Маршрут для создания сообщений
	r.Post("/messages", func(w http.ResponseWriter, r *http.Request) {
		var msg models.Message

		// Декодируем тело запроса в структуру Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Неверный формат данных: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Сохраняем сообщение в базу данных
		if err := database.DB.Create(&msg).Error; err != nil {
			http.Error(w, "Ошибка сохранения сообщения: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем сообщение в Kafka
		err := services.SendMessage("message_topic", msg.Content)
		if err != nil {
			http.Error(w, "Ошибка отправки сообщения в Kafka: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Помечаем сообщение как обработанное
		msg.Processed = true
		if err := database.DB.Save(&msg).Error; err != nil {
			http.Error(w, "Ошибка обновления статуса сообщения: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Возвращаем сохраненное сообщение в ответе
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			http.Error(w, "Ошибка кодирования ответа: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Маршрут для получения статистики обработанных сообщений
	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		var count int64

		// Считаем количество обработанных сообщений
		if err := database.DB.Model(&models.Message{}).Where("processed = ?", true).Count(&count).Error; err != nil {
			http.Error(w, "Ошибка получения статистики: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Возвращаем статистику
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]int64{"processed_messages": count}); err != nil {
			http.Error(w, "Ошибка кодирования ответа: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return r
}
