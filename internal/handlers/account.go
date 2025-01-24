package handlers

import (
	"bank_parser_backend_go/internal/config"
	con "bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DelAccountProfileDirHandler(ac *con.AccountController) gin.HandlerFunc {
	return func(c *gin.Context) {
		ac.DelAccountProfileDir(c)
	}
}

func AuthAccountHandler(ac *con.AccountController, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ac.AuthAccount(c, cfg)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}
