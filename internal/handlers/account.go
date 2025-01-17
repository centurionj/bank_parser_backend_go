package handlers

import (
	con "bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAccountHandler(ac *con.AccountController) gin.HandlerFunc {
	return func(c *gin.Context) {
		account, _ := ac.GetAccount(c)
		c.JSON(http.StatusOK, &account)
	}
}
