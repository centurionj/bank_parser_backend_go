package handlers

import (
	con "bank_parser_backend_go/internal/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ParsTransactionsHandler(tc *con.TransactionController) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Вызываем ParsTransactions и проверяем ошибку
		//if err := tc.ParsTransactions(c); err != nil {
		//	// Если произошла ошибка, ничего не делаем, так как ParsTransactions уже отправил ответ
		//	return
		//}
		tc.ParsTransactions(c)
		// Если всё успешно, отправляем пустой JSON-ответ со статусом 200
		c.JSON(http.StatusOK, nil)
	}
}
