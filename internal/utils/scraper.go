package utils

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
)

// Извлекает данные из HTML страницы

//func extractTransactions(chromeCtx context.Context, accountID int) ([]map[string]interface{}, error) {
//	// Клик по кнопке "Пополнение" и ожидание загрузки контента
//	if err := chromedp.Run(chromeCtx,
//		chromedp.Evaluate(`
//        (() => {
//            const buttons = document.querySelectorAll('button.base-tag__component--CWYoD');
//            for (const button of buttons) {
//                if (button.textContent.trim() === 'Пополнение') {
//                    const clickEvent = new MouseEvent('click', { bubbles: true, cancelable: true });
//                    button.dispatchEvent(clickEvent);
//                    return;
//                }
//            }
//            throw new Error('Button not found');
//        })()
//    `, nil),
//		chromedp.Sleep(RandomDuration(2, 3)), // Ждем после клика для загрузки контента
//	); err != nil {
//		return nil, fmt.Errorf("failed to switch to 'Пополнение': %w", err)
//	}
//
//	// Находим первый блок "Сегодня" и все кнопки внутри него
//	var rawButtons []string
//	if err := chromedp.Run(chromeCtx,
//		chromedp.Sleep(RandomDuration(1, 3)),
//		chromedp.Evaluate(`
//            (() => {
//                const todaySections = Array.from(document.querySelectorAll('div.operations-history-list__section--xX794'));
//                const todaySection = todaySections.find(section =>
//                    section.querySelector('.operation-header__day--Cobpl')?.textContent.trim().toLowerCase() === 'сегодня'
//                );
//                if (!todaySection) {
//                    throw new Error('Today section not found');
//                }
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
//	// Создаем слайс для хранения результатов
//	var results []map[string]interface{}
//
//	// Проходим по каждой кнопке, кликаем на неё, извлекаем HTML-код окна и закрываем его
//	totalButtons := len(todaySectionButtons) // Количество кнопок в блоке "Сегодня"
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
//			chromedp.Sleep(RandomDuration(1, 3)), // Добавляем задержку перед кликом
//			chromedp.Click(fmt.Sprintf(`div.operations-history-list__section--xX794 button[data-test-id="operation-cell"]:nth-of-type(%d)`, buttonIndex), chromedp.NodeVisible),
//		)
//		if err != nil {
//			fmt.Printf("Failed to click on button %d: %v\n", buttonIndex, err)
//			continue // Пропускаем эту кнопку, если клик не удался
//		}
//
//		// Извлечение HTML-кода открывшегося окна
//		var windowHTML string
//		err = chromedp.Run(chromeCtx,
//			chromedp.Sleep(RandomDuration(1, 3)),                                        // Даем время на загрузку попапа
//			chromedp.WaitVisible(`.alfa-components__content--kBgpD`, chromedp.ByQuery),  // Ждем видимости основного контейнера
//			chromedp.WaitVisible(`.financial-analytics__cell--JF7Ay`, chromedp.ByQuery), // Ждем загрузки financial analytics
//			chromedp.WaitVisible(`.details-content__body--zwZ7L`, chromedp.ByQuery),     // Ждем загрузки деталей операции
//			chromedp.Evaluate(`
//				(() => {
//					const popup = document.querySelector('.alfa-components__content--kBgpD');
//					return popup ? popup.outerHTML : '';
//				})()
//			`, &windowHTML),
//		)
//		if err != nil || windowHTML == "" {
//			fmt.Printf("Failed to extract popup HTML content after clicking button %d: %v\n", buttonIndex, err)
//			continue // Пропускаем эту кнопку, если HTML не удалось получить
//		}
//
//		// Создаем запись для текущей транзакции
//		result := map[string]interface{}{
//			"account_id":         accountID,
//			"transaction_number": buttonIndex,
//			"source_code":        windowHTML,
//		}
//		results = append(results, result)
//
//		// Закрытие окна через JavaScript
//		err = chromedp.Run(chromeCtx,
//			chromedp.Sleep(RandomDuration(1, 3)), // Добавляем задержку перед закрытием
//			chromedp.Evaluate(`
//			  (() => {
//			      const closeButton = document.querySelector('button[aria-label="закрыть"]');
//			      if (closeButton) {
//			          const clickEvent = new MouseEvent('click', { bubbles: true, cancelable: true });
//			          closeButton.dispatchEvent(clickEvent);
//			      } else {
//			          console.warn('Close button not found');
//			      }
//			  })()
//			`, nil),
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
//	return results, nil
//}

func extractTransactions(chromeCtx context.Context, accountID int) ([]map[string]interface{}, error) {
	// Клик по кнопке "Пополнение" и ожидание загрузки контента
	if err := chromedp.Run(chromeCtx,
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
		chromedp.Sleep(RandomDuration(2, 3)), // Ждем после клика для загрузки контента
	); err != nil {
		return nil, fmt.Errorf("failed to switch to 'Пополнение': %w", err)
	}

	// Находим первый блок "Сегодня" и все кнопки внутри него
	var rawButtons []string
	if err := chromedp.Run(chromeCtx,
		chromedp.Sleep(RandomDuration(1, 3)),
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

	// Создаем слайс для хранения результатов
	var results []map[string]interface{}
	totalButtons := len(rawButtons)

	for i, _ := range rawButtons {
		buttonIndex := i + 1

		// Проверяем, не отменён ли контекст
		if chromeCtx.Err() != nil {
			fmt.Println("Context canceled. Stopping execution.")
			break
		}

		// Клик по кнопке
		err := chromedp.Run(chromeCtx,
			chromedp.Sleep(RandomDuration(1, 3)),
			chromedp.Click(fmt.Sprintf(`div.operations-history-list__section--xX794 button[data-test-id="operation-cell"]:nth-of-type(%d)`, buttonIndex), chromedp.NodeVisible),
		)
		if err != nil {
			fmt.Printf("Failed to click on button %d: %v\n", buttonIndex, err)
			continue
		}

		// Извлечение данных из попапа
		var transactionData map[string]string
		err = chromedp.Run(chromeCtx,
			chromedp.Sleep(RandomDuration(1, 3)),
			chromedp.WaitVisible(`.alfa-components__content--kBgpD`, chromedp.ByQuery),
			chromedp.Evaluate(`
                (() => {
                    const popup = document.querySelector('.alfa-components__content--kBgpD');
                    if (!popup) return {};
                    
                    // Сумма
                    const amountElement = popup.querySelector('.operation-icon-field__amount--MTKDt');
                    const amount = amountElement ? amountElement.textContent.trim() : '';
                    
                    // Имя автора перевода или "Перевод с карты *..."
                    const authorElement = popup.querySelector('.header__title--BO1ek');
                    const author = authorElement ? authorElement.textContent.trim() : 'Перевод с карты *...';
                    
                    // Категория
                    const categoryElement = popup.querySelector('[data-test-id="category-name-text_content"]');
                    const category = categoryElement ? categoryElement.textContent.trim() : '';
                    
                    // Текущий счёт
                    const accountElements = Array.from(popup.querySelectorAll('[data-test-id="field-with-copy-value-id-text_content"]'));
                    const accountElement = accountElements.find(el => el.textContent.includes("Текущий счёт"));
                    const accountNumber = accountElement ? accountElement.textContent.match(/··(\d{4})/)?.[1] || '' : '';
                    
                    // Дата
                    const dateElement = popup.querySelector('.header__time--PzXMr');
                    const date = dateElement ? dateElement.textContent.trim() : '';
                    
                    return {
                        amount: amount,
                        author: author,
                        category: category,
                        accountNumber: accountNumber,
                        date: date,
                    };
                })()
            `, &transactionData),
		)
		if err != nil || len(transactionData) == 0 {
			fmt.Printf("Failed to extract data after clicking button %d: %v\n", buttonIndex, err)
			continue
		}

		// Создаем запись для текущей транзакции
		result := map[string]interface{}{
			"account_id":         accountID,
			"transaction_number": buttonIndex,
			"amount":             transactionData["amount"],
			"author":             transactionData["author"],
			"category":           transactionData["category"],
			"account_number":     transactionData["accountNumber"],
			"date":               transactionData["date"],
		}
		results = append(results, result)

		// Закрытие окна через JavaScript
		err = chromedp.Run(chromeCtx,
			chromedp.Sleep(RandomDuration(1, 3)),
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
		)
		if err != nil {
			fmt.Printf("Failed to close popup after clicking button %d: %v\n", buttonIndex, err)
			continue
		}

		// Проверяем, является ли текущая кнопка последней
		if buttonIndex == totalButtons {
			fmt.Println("All buttons from 'Today' section have been processed. Exiting.")
			break
		}
	}

	return results, nil
}
