package controllers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	schem "bank_parser_backend_go/internal/schemas"
	"bank_parser_backend_go/internal/utils"
	"errors"
	"fmt"
	cu "github.com/Davincible/chromedp-undetected"
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

// Авторизация аккаунта

func (ac *AccountController) AuthAccount(c *gin.Context, cfg config.Config) error {
	account, err := ac.GetAccount(c)
	if err != nil {
		return err
	}

	loginURL := ac.Cfg.AlphaLoginUrl
	if loginURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing AlphaLoginUrl in config"})
		account.IsErrored = true
		ac.DB.Save(&account)
		return errors.New("Missing AlphaLoginUrl in config")
	}

	conf, err := utils.SetupChromeDriver(c, *account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup ChromeDriver"})
		return errors.New("Failed to setup ChromeDriver")
	}

	ctx, cancel, err := cu.New(conf)
	if err != nil {
		account.IsErrored = true
		ac.DB.Save(&account)
		panic(err)
	}
	defer cancel()

	props := schem.AccountProperties{}
	if props, err = utils.InjectJSProperties(ctx, *account); err != nil {
		log.Printf("Error injecting JS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inject JS properties"})
		account.IsErrored = true
		ac.DB.Save(&account)
		return err
	}

	rand.Seed(time.Now().UnixNano())

	err = chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),

		chromedp.Evaluate(`document.querySelector("button[data-test-id='phone-auth-button']").click()`, nil),
		chromedp.WaitVisible(`input[data-test-id='phoneInput']`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.SendKeys(`input[data-test-id='phoneInput']`, account.PhoneNumber[1:]),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.WaitVisible(`button.phone-auth-browser__submit-button`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.Click(`button.phone-auth-browser__submit-button`, chromedp.NodeVisible),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.WaitVisible(`input[data-test-id='card-account-input']`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.SendKeys(`input[data-test-id='card-account-input']`, account.CardNumber),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.Click(`button[data-test-id='card-account-continue-button']`, chromedp.NodeVisible),
		chromedp.Sleep(time.Duration(rand.Intn(5)+1)*time.Second),
	)

	if err != nil {
		log.Printf("Authorization error: %v", err)
		account.IsErrored = true
		ac.DB.Save(&account)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization failed"})
		return err
	}

	otpCode := account.TemporaryCode
	if otpCode == nil || *otpCode == "" {
		startTime := time.Now()
		timeout := 1 * time.Minute

		for {
			if time.Since(startTime) > timeout {
				log.Println("Timeout waiting for OTP code")
				account.IsErrored = true
				ac.DB.Save(&account)
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Timeout waiting for OTP code"})
				return err
			}

			if err := ac.DB.First(&account, account.ID).Error; err != nil {
				log.Printf("Database error: %v", err)
				account.IsErrored = true
				ac.DB.Save(&account)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return err
			}

			if account.TemporaryCode != nil && *account.TemporaryCode != "" {
				otpCode = account.TemporaryCode
				break
			}

			log.Println("OTP code is still empty, retrying...")
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		}
	}
	time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

	for index, digit := range *otpCode {
		err = chromedp.Run(ctx,
			chromedp.Click(fmt.Sprintf(`input.code-input__input_fq4wa:nth-of-type(%d)`, index+1), chromedp.NodeVisible),
			chromedp.SendKeys(fmt.Sprintf(`input.code-input__input_fq4wa:nth-of-type(%d)`, index+1), string(digit)),
		)
		if err != nil {
			log.Printf("Error entering OTP digit: %v", err)
			account.IsErrored = true
			ac.DB.Save(&account)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error entering OTP digit"})
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}

	cookies, err := utils.GetSessionCookies(c)
	if err != nil {
		log.Printf("Error retrieving session cookies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session cookies"})
		account.IsErrored = true
		ac.DB.Save(&account)
		return err
	}

	if err := ac.saveAccount(account, props, cookies); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save account"})
		account.IsErrored = true
		ac.DB.Save(&account)
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Authorization successful"})
	return nil
}
