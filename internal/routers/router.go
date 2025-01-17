package routers

import (
	"bank_parser_backend_go/internal/config"
	controllers "bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Настройка роутов

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	accountController := controllers.NewAccountController(db, cfg)

	r.POST("api/v1/account/info/", accountController.GetAccountHandler)
	r.POST("api/v1/account/auth/", accountController.AuthAccount)
	r.POST("/api/v1/account/delete_profile/", accountController.DelAccountProfileDirHandler)
}
