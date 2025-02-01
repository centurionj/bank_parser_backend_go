package utils

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"time"
)

// Извлекает данные из HTML страницы

func ExtractTransactions(chromeCtx context.Context) ([]map[string]string, error) {
	var transactions []map[string]string

	// Перезагрузка страницы и выполнение JavaScript для клика по кнопке "Пополнение"
	err := chromedp.Run(chromeCtx,
		chromedp.Reload(),
		chromedp.WaitVisible(`button.base-tag__component--Odrwf span`, chromedp.ByQuery),
		chromedp.Sleep(RandomDuration(2, 4)), // Даем время на загрузку контента
		chromedp.Evaluate(`
            (() => {
                // Находим все кнопки с классом base-tag__component--Odrwf
                const buttons = document.querySelectorAll('button.base-tag__component--Odrwf');
                // Ищем кнопку с текстом "Пополнение"
                for (const button of buttons) {
                    if (button.textContent.includes('Пополнение')) {
                        button.click();
                        return;
                    }
                }
                throw new Error('Button not found');
            })()
        `, nil),
		chromedp.Sleep(RandomDuration(3, 5)), // Ждем после клика
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reload page or switch to 'Пополнение': %w", err)
	}

	// Находим первый элемент operations-history-list__section--epmef
	var firstSection string
	err = chromedp.Run(chromeCtx,
		chromedp.Evaluate(`document.querySelector('div.operations-history-list__section--epmef')?.outerHTML`, &firstSection),
	)
	if err != nil || firstSection == "" {
		return nil, fmt.Errorf("failed to find the first operations-history-list__section--epmef: %w", err)
	}

	// Находим все кнопки operation-cell внутри первого sections
	var operationCells []*cdp.Node
	err = chromedp.Run(chromeCtx,
		chromedp.Nodes(`div.operations-history-list__section--epmef button[data-test-id="operation-cell"]`, &operationCells, chromedp.ByQueryAll),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find operation cells: %w", err)
	}

	// Обработка каждой операции
	for i, cell := range operationCells {
		var transactionData map[string]string = make(map[string]string)

		// Клик по кнопке операции
		err = chromedp.Run(chromeCtx,
			chromedp.Click(fmt.Sprintf(`%s`, cell.NodeValue), chromedp.NodeVisible),
			chromedp.Sleep(RandomDuration(1, 2)), // Ждем загрузки данных
		)
		if err != nil {
			return nil, fmt.Errorf("failed to click on operation cell: %w", err)
		}

		// Получаем HTML-код открытой страницы
		var htmlContent string
		err = chromedp.Run(chromeCtx,
			chromedp.Evaluate(`document.querySelector('div[data-test-id="operation-details-side-panel-body"]')?.outerHTML`, &htmlContent),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to extract HTML content: %w", err)
		}

		// Сохраняем данные
		transactionData["button_number"] = fmt.Sprintf("%d", i+1) // Номер кнопки (начинается с 1)
		transactionData["html_content"] = htmlContent             // HTML-код страницы

		transactions = append(transactions, transactionData)
	}

	return transactions, nil
}

// Парсинг

func FindTransactions(ctx context.Context, cfg config.Config, account *models.Account) error {
	loginURL := cfg.AlphaUrl
	if loginURL == "" {
		return fmt.Errorf("missing AlphaUrl in config")
	}

	// Создаем общий контекст с таймаутом из переданного контекста
	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, time.Duration(cfg.ParserTimeOutSecond)*time.Second)
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
		chromedp.Sleep(RandomDuration(1, 3)),
		//chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to History tab: %w", err)
	}

	transactions, err := ExtractTransactions(chromeCtx)

	if err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}
	println(transactions)

	// Проверка отмены контекста после каждого шага
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
