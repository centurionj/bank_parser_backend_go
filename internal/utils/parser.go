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

func SendTransactions(results string, targetURL string) error {
	// Проверяем, что URL указан
	if targetURL == "" {
		return fmt.Errorf("missing target URL")
	}

	// Формируем полный URL для запроса
	finalUrl := targetURL + "/payments/prepare/"

	// Создаем cookie jar для хранения cookies (если необходимо)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Создаем HTTP клиент с поддержкой cookies
	client := &http.Client{
		Jar: jar,
	}

	// Преобразуем results из строки JSON обратно в map[string]interface{}
	var parsedResults map[string]interface{}
	err = json.Unmarshal([]byte(results), &parsedResults)
	if err != nil {
		return fmt.Errorf("failed to unmarshal results: %w", err)
	}

	payload := map[string]interface{}{
		"account_id": parsedResults["account_id"], // Извлекаем account_id
		"operations": parsedResults["operations"], // Извлекаем operations
	}

	// Преобразуем payload в JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload to JSON: %w", err)
	}

	// Форматируем JSON для красивого вывода
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	// Создаем POST-запрос с правильным Content-Type
	postReq, err := http.NewRequest("POST", finalUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}

	// Устанавливаем заголовок Content-Type
	postReq.Header.Set("Content-Type", "application/json")

	// Отправляем POST-запрос
	postResp, err := client.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer postResp.Body.Close()

	// Читаем тело ответа для логирования или анализа ошибок
	body, err := io.ReadAll(postResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Проверяем статус ответа
	if postResp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST request failed with status %d: %s", postResp.StatusCode, string(body))
	}

	// Выводим успешный результат
	fmt.Printf("POST request sent successfully. Response: %s\n", string(body))

	return nil
}

// Парсинг

func FindTransactions(ctx context.Context, cfg config.Config, account *models.Account) error {
	loginURL := cfg.AlphaUrl
	if loginURL == "" {
		return fmt.Errorf("missing AlphaUrl in config")
	}

	// Создаем контекст с таймаутом для всей операции
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, time.Duration(cfg.ParserTimeOutSecond)*time.Second)
	defer cancelTimeout()

	// Устанавливаем ChromeDriver
	chromeCtx, cancelChrome, err := SetupChromeDriver(timeoutCtx, *account, cfg)
	if err != nil {
		if cancelChrome != nil {
			cancelChrome() // Закрываем Chrome при ошибке
		}
		return fmt.Errorf("failed to setup ChromeDriver: %w", err)
	}
	defer cancelChrome() // Закрываем Chrome при выходе из функции

	// Инъекция JS-свойств
	if _, err := InjectJSProperties(chromeCtx, *account); err != nil {
		return fmt.Errorf("failed to inject JS properties: %w", err)
	}

	// Переход в лк банка
	if err := chromedp.Run(chromeCtx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(RandomDuration(1, 2)),
	); err != nil {
		return fmt.Errorf("failed to navigate to History tab: %w", err)
	}

	// Извлечение транзакций
	result, err := scrapTransactions(chromeCtx, int(account.ID), *account.AccountNumber)
	if err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	// Отправка транзакций
	if err := SendTransactions(result, cfg.BasePythonApiUrl); err != nil {
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
