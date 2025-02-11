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

func (ac *AccountController) DelAccountProfileDir(c *gin.Context) {
	var request schem.AccountIDRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	res, err := utils.ClearSessionDir(int64(request.AccountID), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": res})
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

func (ac *AccountController) accountHandleError(c *gin.Context, account *models.Account, statusCode int, message string, err error) error {
	log.Printf("%s: %v", message, err)
	account.IsErrored = true
	ac.DB.Save(&account)
	c.JSON(statusCode, gin.H{"error": message})
	return err
}

func (ac *AccountController) performLogin(ctx context.Context, account *models.Account, loginURL string) error {
	err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(utils.RandomDuration(1, 3)),
		chromedp.Evaluate(`document.querySelector("button[data-test-id='phone-auth-button']").click()`, nil),
		chromedp.WaitVisible(`input[data-test-id='phoneInput']`, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("login navigation failed: %w", err)
	}

	// Ввод номера телефона
	if err := utils.EnterDigits(ctx, `input[data-test-id='phoneInput']`, account.PhoneNumber[1:]); err != nil {
		return fmt.Errorf("error entering phone number: %w", err)
	}

	return chromedp.Run(ctx,
		chromedp.Click(`button.phone-auth-browser__submit-button`, chromedp.NodeVisible),
		chromedp.WaitVisible(`input[data-test-id='card-account-input']`, chromedp.ByQuery),
	)
}

func (ac *AccountController) enterCardNumber(ctx context.Context, cardNumber string) error {
	if err := utils.EnterDigits(ctx, `input[data-test-id='card-account-input']`, cardNumber); err != nil {
		return fmt.Errorf("error entering card number: %w", err)
	}

	return chromedp.Run(ctx,
		chromedp.Sleep(utils.RandomDuration(1, 3)),
		chromedp.Click(`button[data-test-id='card-account-continue-button']`, chromedp.NodeVisible),
		chromedp.Sleep(utils.RandomDuration(1, 5)),
	)
}

func (ac *AccountController) waitForOTPCode(account *models.Account, timeOut int) (string, error) {
	startTime := time.Now()
	for {
		if time.Since(startTime) > time.Duration(timeOut)*time.Second {
			return "", errors.New("timeout waiting for OTP code")
		}

		if err := ac.DB.First(&account, account.ID).Error; err != nil {
			return "", fmt.Errorf("database error: %w", err)
		}

		if account.TemporaryCode != nil && *account.TemporaryCode != "" {
			return *account.TemporaryCode, nil
		}

		time.Sleep(utils.RandomDuration(1, 5))
	}
}

// Ввод OTP ключа

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

func (ac *AccountController) generateAccountPassword() string {
	rand.Seed(time.Now().UnixNano())
	length := rand.Intn(5) + 4 // Длина от 4 до 8
	password := make([]byte, length)
	for i := range password {
		password[i] = '0' + byte(rand.Intn(10)) // Генерация цифры
	}
	return string(password)
}

// Ввод постоянного пароля для входа в лк

func (ac *AccountController) enterPassword(ctx context.Context, password string) error {
	// Нажимаем на кнопку доверять
	err := chromedp.Run(ctx,
		chromedp.Sleep(utils.RandomDuration(1, 8)),
		chromedp.WaitVisible(`button[data-test-id="trust-device-page-submit-btn"]`, chromedp.ByQuery),
		chromedp.Click(`button[data-test-id="trust-device-page-submit-btn"]`, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("failed to click 'Trust' button: %w", err)
	}

	if err := utils.EnterDigits(ctx, `input[data-test-id="new-password"]`, password); err != nil { // Первичный ввод пароля
		return fmt.Errorf("failed to enter new password: %w", err)
	}

	_ = chromedp.Run(ctx,
		chromedp.Sleep(utils.RandomDuration(1, 3)),
		chromedp.WaitVisible(`input[data-test-id="new-password-again"]`, chromedp.ByQuery),
		chromedp.Click(`input[data-test-id="new-password-again"]`, chromedp.ByQuery),
	)

	if err := utils.EnterDigits(ctx, `input[data-test-id="new-password-again"]`, password); err != nil { // Исправить селектор // Повторный ввод пароля
		return fmt.Errorf("failed to enter password confirmation: %w", err)
	}

	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`button[data-test-id="submit-button"]`, chromedp.ByQuery),
		chromedp.Click(`button[data-test-id="submit-button"]`, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("failed to click 'Save' button: %w", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Sleep(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("setting password failed: %w", err)
	}

	return nil
}

// Авторизация аккаунта

func (ac *AccountController) AuthAccount(c *gin.Context, cfg config.Config) error {
	account, err := ac.GetAccount(c)
	if err != nil {
		return err
	}

	loginURL := ac.Cfg.AlphaUrl
	if loginURL == "" {
		return ac.accountHandleError(c, account, http.StatusInternalServerError, "Missing AlphaUrl in config", errors.New("missing AlphaUrl in config"))
	}

	// Создаем общий контекст с таймаутом
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(cfg.AuthTimeOutSecond)*time.Second)
	defer cancelTimeout()

	// Настраиваем ChromeDriver
	chromeCtx, cancelChrome, err := utils.SetupChromeDriver(timeoutCtx, *account, cfg)
	if err != nil {
		return ac.accountHandleError(c, account, http.StatusInternalServerError, "Failed to setup ChromeDriver", err)
	}
	defer cancelChrome() // Убедимся, что браузер будет закрыт в любом случае

	errChan := make(chan error, 1)

	go func() {
		defer func() {
			// Убедимся, что браузер закроется в случае паники
			if r := recover(); r != nil {
				cancelChrome()
				errChan <- fmt.Errorf("panic occurred: %v", r)
			}
		}()

		// Инъекция JS-свойств
		props, err := utils.InjectJSProperties(chromeCtx, *account)
		if err != nil {
			errChan <- ac.accountHandleError(c, account, http.StatusInternalServerError, "Failed to inject JS properties", err)
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
			errChan <- ac.accountHandleError(c, account, http.StatusRequestTimeout, "Timeout waiting for OTP code", err)
			return
		}

		// Ввод OTP-кода
		if err := ac.enterOTPCode(chromeCtx, otpCode); err != nil {
			errChan <- err
			return
		}

		// Генерация постоянного пароля и его ввод
		password := ac.generateAccountPassword()
		if err := ac.enterPassword(chromeCtx, password); err != nil {
			errChan <- err
			return
		}

		// Получение cookies
		cookies, err := utils.GetSessionCookies(chromeCtx)
		if err != nil {
			errChan <- ac.accountHandleError(c, account, http.StatusInternalServerError, "Failed to retrieve session cookies", err)
			return
		}

		// Сохранение данных аккаунта
		if err := ac.saveAccount(account, props, &cookies); err != nil {
			errChan <- ac.accountHandleError(c, account, http.StatusInternalServerError, "Failed to save account", err)
			return
		}

		// Успешный ответ
		account.Password = &password
		account.IsErrored = false
		ac.DB.Save(&account)
		c.JSON(http.StatusOK, gin.H{"message": "Authorization successful"})
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		// Возвращаем ошибку, если она произошла
		return err
	case <-timeoutCtx.Done():
		// Таймаут: закрываем браузер и возвращаем ошибку
		cancelChrome()
		return ac.accountHandleError(c, account, http.StatusGatewayTimeout, "Operation timed out", errors.New("operation timed out"))
	}
}
