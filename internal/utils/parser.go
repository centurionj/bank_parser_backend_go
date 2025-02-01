package utils

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"time"
)

func FindTransactions(ctx context.Context, cfg config.Config, account *models.Account) error {
	loginURL := cfg.AlphaUrl
	if loginURL == "" {
		return fmt.Errorf("missing AlphaUrl in config")
	}

	// Создаем общий контекст с таймаутом из переданного контекста
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, time.Duration(cfg.AuthTimeOutSecond)*time.Minute)
	defer cancelTimeout()

	// Настраиваем ChromeDriver
	_, chromeCtx, cancelChrome, err := SetupChromeDriver(timeoutCtx, *account, cfg)
	if err != nil {
		cancelChrome() // Убедитесь, что браузер закрыт при ошибке
		return fmt.Errorf("failed to setup ChromeDriver: %w", err)
	}
	defer cancelChrome()

	// Инъекция JS-свойств
	_, err = InjectJSProperties(chromeCtx, *account)
	if err != nil {
		return fmt.Errorf("failed to inject JS properties: %w", err)
	}

	// Начало парсинга (переход на страницу с историей)
	err = chromedp.Run(chromeCtx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(RandomDuration(1, 3)),
	)
	if err != nil {
		return fmt.Errorf("login navigation failed: %w", err)
	}

	err = chromedp.Run(chromeCtx,
		chromedp.WaitVisible(`li[data-test-id='item'] a[href='/history/']`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector("li[data-test-id='item'] a[href='/history/']").click()`, nil),
		chromedp.Sleep(time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to History tab: %w", err)
	}

	// Проверка отмены контекста после каждого шага
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
