package utils

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"math/rand"
)

func generateID64() string {
	id := make([]byte, 8)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(id)
}

func createHeaders(cookies []*network.Cookie) map[string]string {
	headers := make(map[string]string)

	for _, o := range cookies {
		switch o.Name {
		case "DEVICE_APP_ID":
			headers["DEVICE-APP-ID"] = o.Value
		case "fgsscw-alfabank-retail":
			headers["X-GIB-FGSSCw-alfabank-retail"] = o.Value
		case "gsscw-alfabank-retail":
			headers["X-GIB-GSSCw-alfabank-retail"] = o.Value
		case "XSRF-TOKEN":
			headers["X-XSRF-TOKEN"] = o.Value
		}
	}
	headers["x-b3-spanid"] = generateID64()
	headers["x-b3-traceid"] = generateID64()

	return headers
}

func scrapTransactions(chromeCtx context.Context, accountID int, accountNumber string) (string, error) {
	var cookies []*network.Cookie
	if err := chromedp.Run(chromeCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, _ = network.GetCookies().Do(ctx)
		return err
	})); err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	headers := createHeaders(cookies)

	rawData := make(map[string]interface{})
	chromedp.Run(chromeCtx,
		chromedp.Evaluate(fmt.Sprintf(
			`
   (async () => {
   try {
    const response = await fetch('https://web.alfabank.ru/api/v1/operations-history/operations', {
     method: 'POST',
     headers: {
      'Content-Type': 'application/json',
      'AO-ORIGIN-APPLICATION-ID': 'newclick-dashboard-ui',
      'Referer': 'https://web.alfabank.ru/accounts/%s',
      'DEVICE-APP-ID': '%s',
      'X-GIB-FGSSCw-alfabank-retail': '%s',
      'X-GIB-GSSCw-alfabank-retail': '%s',
      'X-XSRF-TOKEN': '%s',
      'x-b3-spanid': '%s',
      'x-b3-traceid': '%s',
      'ZONE-OFFSET': '+03:00'
     },
     credentials: 'include',
     body: '{\"size\":20,\"page\":1,\"filters\":[{\"values\":[\"%s\"],\"type\":\"accounts\"}],\"forced\":true}'
    });
    const data = await response.json();
    return data;
   } catch (err) {
                return JSON.stringify({ error: err.message }); 
            }
   })()
  `, accountNumber, headers["DEVICE-APP-ID"], headers["X-GIB-FGSSCw-alfabank-retail"], headers["X-GIB-GSSCw-alfabank-retail"], headers["X-XSRF-TOKEN"], headers["x-b3-spanid"], headers["x-b3-traceid"], accountNumber), &rawData, func(ep *runtime.EvaluateParams) *runtime.EvaluateParams {
			return ep.WithAwaitPromise(true)
		}),
	)

	operations, ok := rawData["operations"].([]interface{})
	if !ok || len(operations) == 0 {
		return "", fmt.Errorf("key 'operations' must be a non-empty array")
	}

	delete(rawData, "pagesInfo")
	rawData["account_id"] = accountID

	rawJson, err := json.Marshal(rawData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	return string(rawJson), nil
}
