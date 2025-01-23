package controllers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	schem "bank_parser_backend_go/internal/schemas"
	"bank_parser_backend_go/internal/utils"
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

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
	var accountIDRequest schem.AccountIDRequest
	if err := c.ShouldBindJSON(&accountIDRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id_request format"})
		return nil, err
	}

	var account models.Account
	if err := ac.DB.First(&account, accountIDRequest.AccountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return nil, err
	}

	return &account, nil
}

// Удаление директории с данными аккаунта в хроме

func (ac *AccountController) DelAccountProfileDir(c *gin.Context) error {
	var request schem.AccountIDRequest

	// Parse the JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return fmt.Errorf("invalid request format: %w", err)
	}

	// Define the base directory where the folders are located
	baseDir := "./chrome-profile/"

	dirPath := filepath.Join(baseDir, fmt.Sprintf("%d", request.AccountID))

	// Attempt to remove the directory
	if err := os.RemoveAll(dirPath); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": fmt.Sprintf("Failed to delete directory for account ID %d: %s",
					request.AccountID, err.Error()),
			},
		)
		return fmt.Errorf("deletion ERROR: %w", err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Directories deleted successfully"})

	return nil
}

// Сохранение аккаунта

func (ac *AccountController) saveAccount(account *models.Account, props schem.AccountProperties, cookies *string) error {
	account.BufferSize = &props.BufferSize
	account.InputChannels = &props.InputChannels
	account.OutputChannels = &props.OutputChannels
	account.Frequency = &props.RandomFrequency
	account.Start = &props.RandomStart
	account.Stop = &props.RandomStop
	account.CPU = &props.CPU
	account.GPU = &props.GPU
	account.DeviceMemory = &props.DeviceMemory
	account.HardwareConcurrency = &props.HardwareConcurrency
	account.ScreenWidth = &props.DeviceProfile.Screen.Width
	account.ScreenHeight = &props.DeviceProfile.Screen.Height
	account.BatteryVolume = &props.BatteryVolume
	account.IsCharging = &props.IsCharging
	account.SessionCookies = cookies
	account.UserAgent = &props.DeviceProfile.UserAgent
	account.Platform = &props.DeviceProfile.Platform
	account.IsAuthenticated = true

	if err := ac.DB.Save(account).Error; err != nil {
		log.Printf("Error saving account: %v", err)
		return fmt.Errorf("failed to save account: %w", err)
	}
	return nil
}

func (ac *AccountController) handleError(c *gin.Context, account *models.Account, statusCode int, message string, err error) error {
	log.Printf("%s: %v", message, err)
	account.IsErrored = true
	ac.DB.Save(&account)
	c.JSON(statusCode, gin.H{"error": message})
	return err
}

func (ac *AccountController) performLogin(ctx context.Context, account *models.Account, loginURL string) error {
	err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(randomDuration(1, 3)),
		chromedp.Evaluate(`document.querySelector("button[data-test-id='phone-auth-button']").click()`, nil),
		chromedp.WaitVisible(`input[data-test-id='phoneInput']`, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("login navigation failed: %w", err)
	}

	if err := ac.enterDigits(ctx, `input[data-test-id='phoneInput']`, account.PhoneNumber[1:]); err != nil {
		return fmt.Errorf("error entering phone number: %w", err)
	}

	return chromedp.Run(ctx,
		chromedp.Click(`button.phone-auth-browser__submit-button`, chromedp.NodeVisible),
		chromedp.WaitVisible(`input[data-test-id='card-account-input']`, chromedp.ByQuery),
	)
}

func (ac *AccountController) enterDigits(ctx context.Context, selector string, digits string) error {
	for _, digit := range digits {
		if err := chromedp.Run(ctx, chromedp.SendKeys(selector, string(digit))); err != nil {
			return fmt.Errorf("error entering digit '%c': %w", digit, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (ac *AccountController) enterCardNumber(ctx context.Context, cardNumber string) error {
	if err := ac.enterDigits(ctx, `input[data-test-id='card-account-input']`, cardNumber); err != nil {
		return fmt.Errorf("error entering card number: %w", err)
	}

	return chromedp.Run(ctx,
		chromedp.Sleep(randomDuration(1, 3)),
		chromedp.Click(`button[data-test-id='card-account-continue-button']`, chromedp.NodeVisible),
		chromedp.Sleep(randomDuration(1, 5)),
	)
}

func (ac *AccountController) waitForOTPCode(account *models.Account, timeOut int) (string, error) {
	startTime := time.Now()
	for {
		if time.Since(startTime) > time.Duration(timeOut)*time.Minute {
			return "", errors.New("timeout waiting for OTP code")
		}

		if err := ac.DB.First(&account, account.ID).Error; err != nil {
			return "", fmt.Errorf("database error: %w", err)
		}

		if account.TemporaryCode != nil && *account.TemporaryCode != "" {
			return *account.TemporaryCode, nil
		}

		time.Sleep(randomDuration(1, 5))
	}
}

func (ac *AccountController) enterOTPCode(ctx context.Context, otpCode string) error {
	for index, digit := range otpCode {
		if err := chromedp.Run(ctx,
			chromedp.Click(fmt.Sprintf(`input.code-input__input_fq4wa:nth-of-type(%d)`, index+1), chromedp.NodeVisible),
			chromedp.SendKeys(fmt.Sprintf(`input.code-input__input_fq4wa:nth-of-type(%d)`, index+1), string(digit)),
		); err != nil {
			return fmt.Errorf("error entering OTP digit '%c': %w", digit, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func randomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min+1)+min) * time.Second
}

// Авторизация аккаунта

func (ac *AccountController) AuthAccount(c *gin.Context, cfg config.Config) error {
	account, err := ac.GetAccount(c)
	if err != nil {
		return err
	}

	loginURL := ac.Cfg.AlphaLoginUrl
	if loginURL == "" {
		return ac.handleError(c, account, http.StatusInternalServerError, "Missing AlphaLoginUrl in config", errors.New("missing AlphaLoginUrl in config"))
	}

	// Создаём общий контекст с таймаутом
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(cfg.AuthTimeOutMinute)*time.Minute)
	defer cancelTimeout()

	// Настраиваем ChromeDriver
	_, chromeCtx, cancel, err := utils.SetupChromeDriver(timeoutCtx, *account, cfg)
	if err != nil {
		return ac.handleError(c, account, http.StatusInternalServerError, "Failed to setup ChromeDriver", err)
	}
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		// Инъекция JS-свойств
		props, err := utils.InjectJSProperties(chromeCtx, *account)
		if err != nil {
			errChan <- ac.handleError(c, account, http.StatusInternalServerError, "Failed to inject JS properties", err)
			return
		}

		// Авторизация
		if err := ac.performLogin(chromeCtx, account, loginURL); err != nil {
			errChan <- err
			return
		}

		// Ввод номера карты
		if err := ac.enterCardNumber(chromeCtx, account.CardNumber); err != nil {
			errChan <- err
			return
		}

		// Ожидание OTP-кода
		otpCode, err := ac.waitForOTPCode(account, cfg.AuthOTPTimeOutSecond)
		if err != nil {
			errChan <- ac.handleError(c, account, http.StatusRequestTimeout, "Timeout waiting for OTP code", err)
			return
		}

		// Ввод OTP-кода
		if err := ac.enterOTPCode(chromeCtx, otpCode); err != nil {
			errChan <- err
			return
		}

		// Получение cookies
		cookies, err := utils.GetSessionCookies(chromeCtx)
		if err != nil {
			errChan <- ac.handleError(c, account, http.StatusInternalServerError, "Failed to retrieve session cookies", err)
			return
		}

		// Сохранение данных аккаунта
		if err := ac.saveAccount(account, props, &cookies); err != nil {
			errChan <- ac.handleError(c, account, http.StatusInternalServerError, "Failed to save account", err)
			return
		}

		// Успешный ответ
		account.IsErrored = false
		ac.DB.Save(&account)
		c.JSON(http.StatusOK, gin.H{"message": "Authorization successful"})
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-timeoutCtx.Done():
		cancelTimeout()
		return ac.handleError(c, account, http.StatusGatewayTimeout, "Operation timed out", errors.New("operation timed out"))
	}
}
