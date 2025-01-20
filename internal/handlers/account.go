package handlers

import (
	con "bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DelAccountProfileDirHandler(ac *con.AccountController) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ac.DelAccountProfileDir(c)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}

func AuthAccountHandler(ac *con.AccountController) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := ac.AuthAccount(c)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}
