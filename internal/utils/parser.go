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
	// Перезагрузка страницы и выполнение JavaScript для клика по кнопке "Пополнение"
	err := chromedp.Run(chromeCtx,
		chromedp.Reload(),
		chromedp.WaitVisible(`button.base-tag__component--Odrwf span`, chromedp.ByQuery),
		chromedp.Sleep(RandomDuration(2, 4)), // Даем время на загрузку контента
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
		chromedp.Sleep(RandomDuration(3, 5)), // Ждем после клика
	)
	if err != nil {
		return fmt.Errorf("failed to reload page or switch to 'Пополнение': %w", err)
	}

	// Находим блок "Сегодня" и все кнопки внутри него
	var todaySectionButtons []*cdp.Node
	err = chromedp.Run(chromeCtx,
		chromedp.Nodes(`div.operations-history-list__section--epmef:has(.operation-header__day--8BOp7) button[data-test-id="operation-cell"]`, &todaySectionButtons, chromedp.ByQueryAll),
	)
	if err != nil {
		return fmt.Errorf("failed to find operation cells in 'Today' section: %w", err)
	}

	// Создаем мапу для хранения HTML-кодов кнопок
	buttonHTMLMap := make(map[int]string)

	// Проходим по каждой кнопке и сохраняем её HTML-код
	for i, _ := range todaySectionButtons {
		var buttonHTML string
		err = chromedp.Run(chromeCtx,
			chromedp.Evaluate(fmt.Sprintf(`
                (() => {
                    const todaySection = document.querySelector('div.operations-history-list__section--epmef:has(.operation-header__day--8BOp7)');
                    if (!todaySection) {
                        throw new Error('Today section not found');
                    }
                    const buttons = todaySection.querySelectorAll('button[data-test-id="operation-cell"]');
                    return buttons[%d]?.outerHTML || '';
                })()
            `, i), &buttonHTML),
		)
		if err != nil {
			return fmt.Errorf("failed to extract HTML content for button %d: %w", i+1, err)
		}

		// Проверяем, что HTML-код не пустой
		if buttonHTML != "" {
			buttonHTMLMap[i+1] = buttonHTML
		}
	}

	// Выводим мапу в консоль
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

	err = ExtractTransactions(chromeCtx)

	if err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}
	//println(transactions)

	// Проверка отмены контекста после каждого шага
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
