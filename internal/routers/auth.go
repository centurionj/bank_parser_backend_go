package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"bank_parser_backend_go/internal/handlers"
	"github.com/gin-gonic/gin"
)

// Настройка роутов авторизации

func setupAccountRoutes(r *gin.Engine, accountController *controllers.AccountController, cfg *config.Config) {
	r.POST("api/v1/account/auth/", handlers.AuthAccountHandler(accountController, *cfg))
	r.POST("/api/v1/account/delete_profile/", handlers.DelAccountProfileDirHandler(accountController))
}
