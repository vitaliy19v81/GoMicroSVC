package main

import (
	"go_microsvc/config"   // Импортируем пакет для загрузки конфигурации
	"go_microsvc/database" // Импортируем пакет для подключения к базе данных
	"go_microsvc/routes"   // Импортируем пакет для маршрутизации HTTP-запросов
	"log"                  // Импортируем пакет для логирования
	"net/http"             // Импортируем пакет для запуска HTTP сервера
)

func main() {
	// Загрузка конфигурации из файла или переменных окружения
	cfg := config.LoadConfig()

	// Подключение к базе данных PostgreSQL с использованием параметров из конфигурации
	err := database.ConnectDB(cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresHost, cfg.PostgresPort)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Запуск HTTP сервера на порту 8080 с использованием маршрутизации
	err = http.ListenAndServe(":8080", routes.Router())
	if err != nil {
		log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
	}
}
