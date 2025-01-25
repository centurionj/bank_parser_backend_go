package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Структура для хранения конфигурации приложения

type Config struct {
	PostgresDSN          string
	HTTPPort             string
	AlphaUrl             string
	AlphaTransactionUrl  string
	GinMode              string
	AuthTimeOutSecond    int
	AuthOTPTimeOutSecond int
}

// Загружает переменные окружения из файла .env и создает конфиг

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file.\t", err)
	}

	cfg := &Config{
		PostgresDSN: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_DB"),
			os.Getenv("SSL_MODE"),
		),
		HTTPPort:             os.Getenv("HTTP_PORT"),
		AlphaUrl:             os.Getenv("ALPHA_LOGIN_URL"),
		AlphaTransactionUrl:  os.Getenv("ALPHA_TRANSACTION_URL"),
		GinMode:              os.Getenv("GIN_MODE"),
		AuthTimeOutSecond:    180,
		AuthOTPTimeOutSecond: 60,
	}

	return cfg, nil
}
