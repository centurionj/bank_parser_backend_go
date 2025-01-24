package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"bank_parser_backend_go/internal/handlers"
	"github.com/gin-gonic/gin"
)

// Настройка роутов авторизации

func setupAccountRoutes(rg *gin.RouterGroup, accountController *controllers.AccountController, cfg *config.Config) {
	rg.POST("/auth/", handlers.AuthAccountHandler(accountController, *cfg))
	rg.POST("/delete_profile/", handlers.DelAccountProfileDirHandler(accountController))
}
