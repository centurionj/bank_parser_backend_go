package handlers

import (
	con "bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ParsTransactionsHandler(tc *con.TransactionController) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tc.ParsTransactions(c)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}
