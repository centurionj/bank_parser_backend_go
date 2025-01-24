package routers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/controllers"
	"bank_parser_backend_go/internal/handlers"
	"github.com/gin-gonic/gin"
)

// Настройка роутов парсинга

func setupParserRoutes(r *gin.Engine, accountController *controllers.AccountController, cfg *config.Config) {
	r.POST("api/v1/parser/auth/", handlers.AuthAccountHandler(accountController, *cfg))
	r.POST("/api/v1/parser/pars/", handlers.DelAccountProfileDirHandler(accountController))
}
