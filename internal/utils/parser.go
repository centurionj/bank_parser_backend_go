package utils

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"time"
)

//// Извлекает данные из HTML страницы
//
//func extractTransactions(chromeCtx context.Context) (map[int]string, error) {
//	// Клик по кнопке "Пополнение" и ожидание загрузки контента
//	if err := chromedp.Run(chromeCtx,
//		chromedp.Reload(),
//		chromedp.WaitVisible(`button.base-tag__component--Odrwf span`, chromedp.ByQuery),
//		chromedp.Sleep(RandomDuration(1, 4)),
//		chromedp.Evaluate(`
//            (() => {
//                const buttons = document.querySelectorAll('button.base-tag__component--Odrwf');
//                for (const button of buttons) {
//                    if (button.textContent.includes('Пополнение')) {
//                        button.click();
//                        return;
//                    }
//                }
//                throw new Error('Button not found');
//            })()
//        `, nil),
//		chromedp.Sleep(RandomDuration(1, 3)),
//	); err != nil {
//		return nil, fmt.Errorf("failed to reload page or switch to 'Пополнение': %w", err)
//	}
//
//	// Находим первый блок "Сегодня" и все кнопки внутри него
//	var rawButtons []string
//	if err := chromedp.Run(chromeCtx,
//		chromedp.Evaluate(`
//            (() => {
//                // Находим первый блок "Сегодня"
//                const todaySections = Array.from(document.querySelectorAll('div.operations-history-list__section--epmef'));
//                const todaySection = todaySections.find(section => section.querySelector('.operation-header__day--8BOp7')?.textContent === 'сегодня');
//                if (!todaySection) {
//                    throw new Error('Today section not found');
//                }
//                // Возвращаем HTML всех кнопок внутри этого блока
//                return Array.from(todaySection.querySelectorAll('button[data-test-id="operation-cell"]'), button => button.outerHTML);
//            })()
//        `, &rawButtons),
//	); err != nil {
//		return nil, fmt.Errorf("failed to find operation cells in 'Today' section: %w", err)
//	}
//
//	// Преобразуем сырые данные в узлы DOM
//	todaySectionButtons := make([]*cdp.Node, len(rawButtons))
//	for i, html := range rawButtons {
//		if html != "" {
//			todaySectionButtons[i] = &cdp.Node{NodeValue: html}
//		}
//	}
//
//	// Создаем мапу для хранения HTML-кодов окон
//	windowHTMLMap := make(map[int]string)
//
//	// Проходим по каждой кнопке, кликаем на неё, извлекаем HTML-код окна и закрываем его
//	totalButtons := len(todaySectionButtons) // Количество кнопок в блоке "Сегодня"
//	println(totalButtons)
//	for i, _ := range todaySectionButtons {
//		buttonIndex := i + 1 // Нумерация кнопок начинается с 1
//
//		// Проверяем, не отменён ли контекст
//		if chromeCtx.Err() != nil {
//			fmt.Println("Context canceled. Stopping execution.")
//			break // Прерываем цикл, если контекст отменён
//		}
//
//		// Клик по кнопке
//		err := chromedp.Run(chromeCtx,
//			chromedp.Click(fmt.Sprintf(`div.operations-history-list__section--epmef:has(.operation-header__day--8BOp7) button[data-test-id="operation-cell"]:nth-of-type(%d)`, buttonIndex), chromedp.NodeVisible),
//			chromedp.Sleep(RandomDuration(1, 4)), // Ждем после клика
//		)
//		if err != nil {
//			fmt.Printf("Failed to click on button %d: %v\n", buttonIndex, err)
//			continue // Пропускаем эту кнопку, если клик не удался
//		}
//
//		// Извлечение HTML-кода открывшегося окна
//		var windowHTML string
//		err = chromedp.Run(chromeCtx,
//			chromedp.Evaluate(`
//                (() => {
//                    const popup = document.querySelector('.content__content--jQ6je');
//                    return popup ? popup.outerHTML : '';
//                })()
//            `, &windowHTML),
//		)
//		if err != nil || windowHTML == "" {
//			fmt.Printf("Failed to extract popup HTML content after clicking button %d: %v\n", buttonIndex, err)
//			continue // Пропускаем эту кнопку, если HTML не удалось получить
//		}
//
//		// Добавляем HTML-код окна в мапу
//		windowHTMLMap[buttonIndex] = windowHTML
//
//		// Закрытие окна через JavaScript
//		err = chromedp.Run(chromeCtx,
//			chromedp.Evaluate(`
//                (() => {
//                    const closeButton = document.querySelector('button[aria-label="закрыть"]');
//                    if (closeButton) {
//                        closeButton.click();
//                    } else {
//                        console.warn('Close button not found');
//                    }
//                })()
//            `, nil),
//			chromedp.Sleep(RandomDuration(1, 3)), // Ждем после закрытия окна
//		)
//		if err != nil {
//			fmt.Printf("Failed to close popup after clicking button %d: %v\n", buttonIndex, err)
//			continue // Пропускаем эту кнопку, если закрытие не удалось
//		}
//
//		// Проверяем, является ли текущая кнопка последней
//		if buttonIndex == totalButtons {
//			fmt.Println("All buttons from 'Today' section have been processed. Exiting.")
//			break // Выход из цикла после обработки последней кнопки
//		}
//	}
//
//	return windowHTMLMap, nil
//}

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
	_, err = extractTransactions(chromeCtx)
	if err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	// Проверка состояния контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
