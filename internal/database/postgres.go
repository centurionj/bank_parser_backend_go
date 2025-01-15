package database

import (
	"bank_parser_backend_go/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Подключение к бд

func ConnectPostgres(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
