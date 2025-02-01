package utils

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"math/rand"
	"time"
)

func RandomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min+1)+min) * time.Second
}

func EnterDigits(ctx context.Context, selector string, digits string) error {
	for _, digit := range digits {
		if err := chromedp.Run(ctx, chromedp.SendKeys(selector, string(digit))); err != nil {
			return fmt.Errorf("error entering digit '%c': %w", digit, err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
