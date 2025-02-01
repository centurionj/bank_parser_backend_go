package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	accountController := controllers.NewAccountController(db, cfg)
	transactionController := controllers.NewTransactionController(db, cfg)

	// Группа для маршрутов аккаунта
	accountGroup := r.Group("api/v1/account")
	setupAccountRoutes(accountGroup, accountController, cfg)

	// Группа для маршрутов парсинга
	parserGroup := r.Group("api/v1/parser")
	setupParserRoutes(parserGroup, transactionController, cfg)
}
