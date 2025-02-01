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

func ExtractTransactions(chromeCtx context.Context) error {
	// Клик по кнопке "Пополнение" и ожидание загрузки контента
	if err := chromedp.Run(chromeCtx,
		chromedp.Reload(),
		chromedp.WaitVisible(`button.base-tag__component--Odrwf span`, chromedp.ByQuery),
		chromedp.Sleep(RandomDuration(2, 4)),
		chromedp.Evaluate(`
            (() => {
                const buttons = document.querySelectorAll('button.base-tag__component--Odrwf');
                for (const button of buttons) {
                    if (button.textContent.includes('Пополнение')) {
                        button.click();
                        return;
                    }
                }
                throw new Error('Button not found');
            })()
        `, nil),
		chromedp.Sleep(RandomDuration(3, 5)),
	); err != nil {
		return fmt.Errorf("failed to reload page or switch to 'Пополнение': %w", err)
	}

	// Поиск кнопок в блоке "Сегодня"
	var todaySectionButtons []*cdp.Node
	if err := chromedp.Run(chromeCtx,
		chromedp.Nodes(`div.operations-history-list__section--epmef:has(.operation-header__day--8BOp7) button[data-test-id="operation-cell"]`, &todaySectionButtons, chromedp.ByQueryAll),
	); err != nil {
		return fmt.Errorf("failed to find operation cells in 'Today' section: %w", err)
	}

	// Извлечение HTML-кодов кнопок
	buttonHTMLMap := make(map[int]string)
	for i := range todaySectionButtons {
		var buttonHTML string
		if err := chromedp.Run(chromeCtx,
			chromedp.Evaluate(fmt.Sprintf(`
                (() => {
                    const button = document.querySelector('div.operations-history-list__section--epmef:has(.operation-header__day--8BOp7) button[data-test-id="operation-cell"]:nth-of-type(%d)');
                    return button ? button.outerHTML : '';
                })()
            `, i+1), &buttonHTML),
		); err != nil {
			return fmt.Errorf("failed to extract HTML content for button %d: %w", i+1, err)
		}
		if buttonHTML != "" {
			buttonHTMLMap[i+1] = buttonHTML
		}
	}

	// Вывод мапы в консоль
	fmt.Println("Button HTML Map:")
	for key, value := range buttonHTMLMap {
		fmt.Printf("Button %d: %s\n", key, value)
	}
	return nil
}

// Парсинг

func FindTransactions(ctx context.Context, cfg config.Config, account *models.Account) error {
	loginURL := cfg.AlphaUrl
	if loginURL == "" {
		return fmt.Errorf("missing AlphaUrl in config")
	}

	// Настройка ChromeDriver
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
	if err := ExtractTransactions(chromeCtx); err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
