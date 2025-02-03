package utils

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// отправляет список транзакций на указанный URL

func SendTransactions(ctx context.Context, results []map[string]interface{}, targetURL string) error {
	// Проверяем, что URL указан
	if targetURL == "" {
		return fmt.Errorf("missing target URL")
	}

	finalUrl := targetURL + "/prser/"

	// Создаем cookie jar для хранения cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Создаем HTTP клиент с поддержкой cookies
	client := &http.Client{
		Jar: jar,
	}

	// Шаг 2: Оборачиваем результаты в нужную структуру
	payload := map[string]interface{}{
		"transactions": results,
	}

	// Преобразуем payload в JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload to JSON: %w", err)
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	fmt.Println("Sending the following JSON payload:")
	fmt.Println(prettyJSON.String())

	postReq, err := http.NewRequest("POST", finalUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}

	// Отправляем POST-запрос
	postResp, err := client.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer postResp.Body.Close()

	// Проверяем статус ответа
	if postResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("POST request failed with status %d: %s", postResp.StatusCode, string(body))
	}

	fmt.Println("POST request sent successfully.")
	return nil
}

// Парсинг

func FindTransactions(ctx context.Context, cfg config.Config, account *models.Account) error {
	loginURL := cfg.AlphaUrl
	if loginURL == "" {
		return fmt.Errorf("missing AlphaUrl in config")
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, time.Duration(cfg.ParserTimeOutSecond)*time.Second)
	defer cancelTimeout()

	_, chromeCtx, cancelChrome, err := SetupChromeDriver(timeoutCtx, *account, cfg)
	if err != nil {
		cancelChrome()
		return fmt.Errorf("failed to setup ChromeDriver: %w", err)
	}
	defer cancelChrome()

	// Инъекция JS-свойств
	if _, err := InjectJSProperties(chromeCtx, *account); err != nil {
		return fmt.Errorf("failed to inject JS properties: %w", err)
	}

	// Переход на страницу истории
	if err := chromedp.Run(chromeCtx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(RandomDuration(1, 3)),
		chromedp.WaitVisible(`li[data-test-id='item'] a[href='/history/']`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector("li[data-test-id='item'] a[href='/history/']").click()`, nil),
		chromedp.Sleep(RandomDuration(1, 3)),
	); err != nil {
		return fmt.Errorf("failed to navigate to History tab: %w", err)
	}

	// Извлечение транзакций
	results, err := extractTransactions(chromeCtx, int(account.ID))

	if err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	if err := SendTransactions(ctx, results, cfg.BasePythonApiUrl); err != nil {
		return fmt.Errorf("failed to send transactions: %w", err)
	}

	// Проверка состояния контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
