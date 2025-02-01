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
		return err
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tc.Cfg.ParserTimeOutSecond)*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	defer close(errChan)

	var account models.Account
	if err := tc.DB.First(&account, accountIDRequest.AccountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errChan <- errors.New("account not found")
		} else {
			errChan <- errors.New("database error")
		}
	}
	go func(accountID int) {
		defer func() {
			// Убедитесь, что браузер закроется в случае паники
			if r := recover(); r != nil {
				cancel()
				errChan <- fmt.Errorf("panic occurred: %v", r)
			}
		}()
		// Передаем контекст с таймаутом в FindTransactions
		err := utils.FindTransactions(ctx, *tc.Cfg, &account)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}(int(account.ID))

	for err := range errChan {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "All transactions parsed successfully"})
	return nil
}
