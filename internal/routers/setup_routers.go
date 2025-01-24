package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	accountController := controllers.NewAccountController(db, cfg)

	// Роуты аккаунта
	setupAccountRoutes(r, accountController, cfg)

	// Роуты парсинга
	setupParserRoutes(r, accountController, cfg)
}
