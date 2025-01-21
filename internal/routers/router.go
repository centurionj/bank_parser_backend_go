package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"bank_parser_backend_go/internal/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Настройка роутов

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	accountController := controllers.NewAccountController(db, cfg)

	r.POST("api/v1/account/auth/", handlers.AuthAccountHandler(accountController, *cfg))
	r.POST("/api/v1/account/delete_profile/", handlers.DelAccountProfileDirHandler(accountController))
}
