package main

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/database"
	"bank_parser_backend_go/internal/server"
	"log"
)

// Точка входа в приложение

func main() {
	// Загрузка конфига
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Подключение к дб
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	// Запуск HTTP сервера
	httpServer := server.NewHTTPServer(cfg, db)
	if err := httpServer.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
