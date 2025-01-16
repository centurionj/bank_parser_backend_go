package controllers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	"fmt"
	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Структура для получения account_id из json

type AccountIDRequest struct {
	AccountID int `json:"account_id"`
}

type AccountController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

// Конструктор контроллеров Account

func NewAccountController(db *gorm.DB, cfg *config.Config) *AccountController {
	return &AccountController{DB: db, Cfg: cfg}
}

// Метод получения Account по его id

func (ac *AccountController) GetAccount(c *gin.Context) (*models.Account, error) {
	var accountIDRequest AccountIDRequest
	if err := c.ShouldBindJSON(&accountIDRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id_request format"})
		return nil, err
	}

	var account models.Account
	if err := ac.DB.First(&account, accountIDRequest.AccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return nil, err
	}

	return &account, nil
}

func (ac *AccountController) GetAccountHandler(c *gin.Context) {
	account, _ := ac.GetAccount(c)

	c.JSON(http.StatusOK, &account)
}

// Метод авторизации в ЛК банка

func (ac *AccountController) AuthAccount(c *gin.Context) {
	account, err := ac.GetAccount(c)
	if err != nil {
		return
	}

	loginURL := ac.Cfg.AlphaLoginUrl
	if loginURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing AlphaLoginUrl in config"})
		return
	}

	conf := cu.NewConfig(
		cu.WithContext(c),
	)

	// Настройки ChromeFlags и других параметров
	conf.ChromeFlags = append(conf.ChromeFlags,
		chromedp.Flag("user-data-dir", "./chrome-profile/"+strconv.Itoa(int(account.ID))),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 OPR/114.0.0.0 (Edition Yx 05)"),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-features", "FontEnumeration"),
		chromedp.ExecPath("/usr/bin/google-chrome"), // Убедитесь, что путь правильный
	)

	ctx, cancel, err := cu.New(conf)
	if err != nil {
		panic(err)
	}
	defer cancel()

	// Выполнение действий в браузере
	err = chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(3*time.Second), // Ждем загрузку страницы
		// Ввод номера телефона
		chromedp.SendKeys(`input[data-test-id='phoneInput']`, account.PhoneNumber[1:]),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`button.phone-auth-browser__submit-button`, chromedp.NodeVisible),
		chromedp.Sleep(2*time.Second),
		// Ввод номера карты
		chromedp.SendKeys(`input[data-test-id='card-account-input']`, account.CardNumber),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`button[data-test-id='card-account-continue-button']`, chromedp.NodeVisible),
		chromedp.Sleep(5*time.Second), // Ожидание ввода одноразового кода
	)

	if err != nil {
		log.Printf("Authorization error: %v", err)
		account.IsErrored = true
		ac.DB.Save(&account)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization failed"})
		return
	}

	// Бесконечный цикл ожидания ввода одноразового кода
	for {
		if err := ac.DB.First(&account, account.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if account.TemporaryCode != nil {
			break
		}

		time.Sleep(5 * time.Second) // Пауза между проверками
	}

	otpCode := *account.TemporaryCode

	// Ввод одноразового кода по символам
	for index, digit := range otpCode {
		err = chromedp.Run(ctx,
			chromedp.SendKeys(fmt.Sprintf(`input.code-input__input_1yhze:nth-child(%d)`, index+1), string(digit)),
		)
		if err != nil {
			log.Printf("Error entering OTP digit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error entering OTP digit"})
			return
		}
	}

	// Получение User-Agent
	var userAgent string
	err = chromedp.Run(ctx, chromedp.Evaluate(`navigator.userAgent`, &userAgent))
	if err != nil {
		log.Printf("Error retrieving User-Agent: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving User-Agent"})
		return
	}

	// Обновляем информацию в базе данных
	account.IsAuthenticated = true
	account.SessionCookies = nil   // Сохраняем cookies сессии, если требуется
	account.UserAgent = &userAgent // Сохраняем User-Agent
	ac.DB.Save(&account)

	c.JSON(http.StatusOK, gin.H{"message": "Authorization successful"})
}
