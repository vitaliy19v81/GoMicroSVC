package models

import (
	"gorm.io/gorm"
)

// Message представляет структуру сообщения в базе данных
// swagger:model
type Message struct {
	gorm.Model        // Включает ID, CreatedAt, UpdatedAt, DeletedAt
	Content    string `json:"content"`
	Processed  bool   `json:"processed" gorm:"default:false"` // Устанавливаем значение по умолчанию для processed
}

// CreateMessageRequest Структура для передачи данных при создании сообщения
// swagger:model CreateMessageRequest
type CreateMessageRequest struct {
	Content string `json:"content"` // validate:"required"`
}

// CreateMessageResponse структура для отображения ответа после создания сообщения
// swagger:model CreateMessageResponse
type CreateMessageResponse struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	Processed bool   `json:"processed"`
}
