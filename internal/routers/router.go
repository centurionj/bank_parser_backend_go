package routers

import (
	controllers "bank_parser_backend_go/internal/controller"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Настройка роутов

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	accountController := controllers.NewAccountController(db)

	// Обработчик на получение данных о аккаунте
	r.POST("api/v1/auth/", accountController.GetAccount)
}
