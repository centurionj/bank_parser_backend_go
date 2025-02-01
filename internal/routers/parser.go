package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"bank_parser_backend_go/internal/handlers"
	"github.com/gin-gonic/gin"
)

// Настройка роутов парсинга

func setupParserRoutes(rg *gin.RouterGroup, transactionController *controllers.TransactionController, cfg *config.Config) {
	rg.POST("/parse/", handlers.ParsTransactionsHandler(transactionController))
}
