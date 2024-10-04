package database

import (
	"fmt"                     // Для форматирования строки подключения (DSN)
	"gorm.io/driver/postgres" // Импортируем драйвер для работы с PostgreSQL
	"gorm.io/gorm"            // Импортируем GORM - ORM для работы с базой данных
	"log"                     // Для логирования ошибок и успешных подключений
)

var DB *gorm.DB // Глобальная переменная для хранения экземпляра подключения к базе данных

// Функция для подключения к базе данных PostgreSQL
func ConnectDB(user, password, dbname, host, port string) error {
	// Формируем строку подключения (DSN) с параметрами базы данных
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	var err error
	// Пытаемся подключиться к базе данных через GORM
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// Проверяем на наличие ошибки при подключении
	if err != nil {
		// Логируем ошибку и возвращаем её наверх
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
		return err
	}

	// Логируем успешное подключение
	log.Println("Успешное подключение к базе данных")
	return nil
}
