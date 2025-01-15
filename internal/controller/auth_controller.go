package controllers

import (
	"bank_parser_backend_go/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AccountController struct {
	DB *gorm.DB
}

func NewAccountController(db *gorm.DB) *AccountController {
	return &AccountController{DB: db}
}

func (ac *AccountController) GetAccount(c *gin.Context) {
	var request struct {
		AccountID int `json:"account_id"`
	}

	// Парсинг JSON запроса
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var account models.Account
	if err := ac.DB.First(&account, request.AccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, account)
}
