package controllers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	schem "bank_parser_backend_go/internal/schemas"
	"bank_parser_backend_go/internal/utils"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type TransactionController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

// Конструктор контроллеров Transaction

func NewTransactionController(db *gorm.DB, cfg *config.Config) *TransactionController {
	return &TransactionController{DB: db, Cfg: cfg}
}

func (tc *TransactionController) ParsTransactions(c *gin.Context) error {
	var accountIDRequest schem.AccountIDRequest
	if err := c.ShouldBindJSON(&accountIDRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return err // Возвращаем ошибку
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tc.Cfg.ParserTimeOutSecond)*time.Second)
	defer cancel()

	var account models.Account
	if err := tc.DB.First(&account, accountIDRequest.AccountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return fmt.Errorf("account not found")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return fmt.Errorf("database error")
	}

	// Канал для обработки ошибок
	errChan := make(chan error, 1)
	defer close(errChan)

	// Запускаем парсинг в отдельной горутине
	go func(accountID int) {
		defer func() {
			// Убедимся, что браузер закрывается при панике
			if r := recover(); r != nil {
				cancel()
				errChan <- fmt.Errorf("panic occurred: %v", r)
			}
		}()
		// Выполняем FindTransactions
		err := utils.FindTransactions(ctx, *tc.Cfg, &account)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil // Успешное завершение
	}(int(account.ID))

	// Ждем завершения горутины
	select {
	case err := <-errChan:
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err // Возвращаем ошибку
		}
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Parsing timeout exceeded"})
			return fmt.Errorf("parsing timeout exceeded")
		}
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "All transactions parsed successfully"})
	return nil // Возвращаем nil при успешном завершении
}
