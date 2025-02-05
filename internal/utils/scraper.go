package utils

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// Извлекает данные из HTML страницы

func extractTransactions(chromeCtx context.Context, accountID int) ([]map[string]interface{}, error) {
	// Клик по кнопке "Пополнение" и ожидание загрузки контента
	if err := chromedp.Run(chromeCtx,
		chromedp.Sleep(RandomDuration(1, 3)), // Даем время на полную загрузку страницы
		chromedp.Evaluate(`
        (() => {
            const buttons = document.querySelectorAll('button.base-tag__component--CWYoD');
            for (const button of buttons) {
                if (button.textContent.trim() === 'Пополнение') {
                    const clickEvent = new MouseEvent('click', { bubbles: true, cancelable: true });
                    button.dispatchEvent(clickEvent);
                    return;
                }
            }
            throw new Error('Button not found');
        })()
    `, nil),
		chromedp.Sleep(RandomDuration(3, 5)), // Ждем после клика для загрузки контента
	); err != nil {
		return nil, fmt.Errorf("failed to switch to 'Пополнение': %w", err)
	}

	// Находим первый блок "Сегодня" и все кнопки внутри него
	var rawButtons []string
	if err := chromedp.Run(chromeCtx,
		chromedp.Sleep(RandomDuration(2, 4)),
		chromedp.Evaluate(`
            (() => {
                const todaySections = Array.from(document.querySelectorAll('div.operations-history-list__section--xX794'));
                const todaySection = todaySections.find(section => 
                    section.querySelector('.operation-header__day--Cobpl')?.textContent.trim().toLowerCase() === 'сегодня'
                );
                if (!todaySection) {
                    throw new Error('Today section not found');
                }
                return Array.from(todaySection.querySelectorAll('button[data-test-id="operation-cell"]'), button => button.outerHTML);
            })()
        `, &rawButtons),
	); err != nil {
		return nil, fmt.Errorf("failed to find operation cells in 'Today' section: %w", err)
	}

	// Преобразуем сырые данные в узлы DOM
	todaySectionButtons := make([]*cdp.Node, len(rawButtons))
	for i, html := range rawButtons {
		if html != "" {
			todaySectionButtons[i] = &cdp.Node{NodeValue: html}
		}
	}

	// Создаем слайс для хранения результатов
	var results []map[string]interface{}

	// Проходим по каждой кнопке, кликаем на неё, извлекаем HTML-код окна и закрываем его
	totalButtons := len(todaySectionButtons) // Количество кнопок в блоке "Сегодня"
	for i, _ := range todaySectionButtons {
		buttonIndex := i + 1 // Нумерация кнопок начинается с 1

		// Проверяем, не отменён ли контекст
		if chromeCtx.Err() != nil {
			fmt.Println("Context canceled. Stopping execution.")
			break // Прерываем цикл, если контекст отменён
		}

		// Клик по кнопке
		err := chromedp.Run(chromeCtx,
			chromedp.Sleep(RandomDuration(1, 3)), // Добавляем задержку перед кликом
			chromedp.Click(fmt.Sprintf(`div.operations-history-list__section--xX794 button[data-test-id="operation-cell"]:nth-of-type(%d)`, buttonIndex), chromedp.NodeVisible),
			chromedp.Sleep(RandomDuration(2, 4)), // Ждем после клика
		)
		if err != nil {
			fmt.Printf("Failed to click on button %d: %v\n", buttonIndex, err)
			continue // Пропускаем эту кнопку, если клик не удался
		}

		// Извлечение HTML-кода открывшегося окна
		var windowHTML string
		err = chromedp.Run(chromeCtx,
			chromedp.Sleep(RandomDuration(1, 3)),                               // Даем время на загрузку попапа
			chromedp.WaitVisible(`.content__content--jQ6je`, chromedp.ByQuery), // Ждем видимости попапа
			chromedp.Evaluate(`
                (() => {
                    const popup = document.querySelector('.content__content--jQ6je');
                    return popup ? popup.outerHTML : '';
                })()
            `, &windowHTML),
		)
		if err != nil || windowHTML == "" {
			fmt.Printf("Failed to extract popup HTML content after clicking button %d: %v\n", buttonIndex, err)
			continue // Пропускаем эту кнопку, если HTML не удалось получить
		}

		// Создаем запись для текущей транзакции
		result := map[string]interface{}{
			"account_id":         accountID,
			"transaction_number": buttonIndex,
			"source_code":        windowHTML,
		}
		results = append(results, result)

		// Закрытие окна через JavaScript
		err = chromedp.Run(chromeCtx,
			chromedp.Sleep(RandomDuration(1, 3)), // Добавляем задержку перед закрытием
			chromedp.Evaluate(`
                (() => {
                    const closeButton = document.querySelector('button[aria-label="закрыть"]');
                    if (closeButton) {
                        const clickEvent = new MouseEvent('click', { bubbles: true, cancelable: true });
                        closeButton.dispatchEvent(clickEvent);
                    } else {
                        console.warn('Close button not found');
                    }
                })()
            `, nil),
			chromedp.Sleep(RandomDuration(1, 3)), // Ждем после закрытия окна
		)
		if err != nil {
			fmt.Printf("Failed to close popup after clicking button %d: %v\n", buttonIndex, err)
			continue // Пропускаем эту кнопку, если закрытие не удалось
		}

		// Проверяем, является ли текущая кнопка последней
		if buttonIndex == totalButtons {
			fmt.Println("All buttons from 'Today' section have been processed. Exiting.")
			break // Выход из цикла после обработки последней кнопки
		}
	}

	return results, nil
}
