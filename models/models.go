package models

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Content   string `json:"content"`
	Processed bool   `json:"processed"`
}
