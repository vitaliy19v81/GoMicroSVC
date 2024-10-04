package crud

import (
	"go_microsvc/database"
)

type User struct {
	ID    uint
	Name  string
	Email string
}

// Создание пользователя
func CreateUser(user User) error {
	result := database.DB.Create(&user)
	return result.Error
}

// Получение всех пользователей
func GetAllUsers() ([]User, error) {
	var users []User
	result := database.DB.Find(&users)
	return users, result.Error
}
