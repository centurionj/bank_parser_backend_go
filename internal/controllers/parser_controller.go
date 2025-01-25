package controllers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/utils"
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (ac *AccountController) ParsTransactions(c *gin.Context, cfg config.Config) error {
	account, err := ac.GetAccount(c)
	if err != nil {
		return err
	}

	loginURL := ac.Cfg.AlphaUrl
	if loginURL == "" {
		return ac.handleError(c, account, http.StatusInternalServerError, "Missing AlphaUrl in config", errors.New("missing AlphaUrl in config"))
	}

	// Создаём общий контекст с таймаутом
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(cfg.AuthTimeOutSecond)*time.Minute)
	defer cancelTimeout()

	// Настраиваем ChromeDriver
	_, chromeCtx, cancelChrome, err := utils.SetupChromeDriver(timeoutCtx, *account, cfg)
	if err != nil {
		return ac.handleError(c, account, http.StatusInternalServerError, "Failed to setup ChromeDriver", err)
	}
	defer cancelChrome()

	// Инъекция JS-свойств
	_, err = utils.InjectJSProperties(chromeCtx, *account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inject JS properties"})
		return err
	}

	err = chromedp.Run(chromeCtx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(5*time.Minute),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login navigation failed"})
		return err
	}

	return nil

}
