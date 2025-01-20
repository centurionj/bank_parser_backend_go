package utils

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	schem "bank_parser_backend_go/internal/schemas"
	"context"
	"encoding/json"
	"fmt"
	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Набор профилей для различных устройств

func getRandomDeviceProfile() schem.DeviceProfile {
	devices := []schem.DeviceProfile{
		{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
			Platform:  "Win32",
			Screen:    struct{ Width, Height int }{1920, 1080},
		},
		{
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
			Platform:  "MacIntel",
			Screen:    struct{ Width, Height int }{1440, 900},
		},
		{
			UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
			Platform:  "Linux x86_64",
			Screen:    struct{ Width, Height int }{1920, 1080},
		},
		{
			UserAgent: "Mozilla/5.0 (Linux; Android 12; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.0.0 Mobile Safari/537.36",
			Platform:  "Linux armv7l",
			Screen:    struct{ Width, Height int }{1080, 2340},
		},
		{
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
			Platform:  "iPhone",
			Screen:    struct{ Width, Height int }{390, 844},
		},
		{
			UserAgent: "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
			Platform:  "iPad",
			Screen:    struct{ Width, Height int }{810, 1080},
		},
	}

	// Выбираем случайный профиль устройства
	return devices[rand.Intn(len(devices))]
}

func createAccountProperties(account models.Account) schem.AccountProperties {
	var props schem.AccountProperties

	if account.IsAuthenticated != true {
		props.DeviceProfile = getRandomDeviceProfile()

		props.RandomFrequency = 8000 + rand.Intn(4000)
		props.RandomStart = fmt.Sprintf("%.2f", rand.Float64()/10)
		props.RandomStop = fmt.Sprintf("%.1f", 0.1+rand.Float64()/10)

		bufferSizes := []int{256, 512, 1024, 2048, 4096, 8192, 16384}
		props.BufferSize = bufferSizes[rand.Intn(len(bufferSizes))]
		props.InputChannels = rand.Intn(2) + 1
		props.OutputChannels = rand.Intn(2) + 1

		gpus := []string{"NVIDIA GeForce GTX 1660", "AMD Radeon RX 580", "Intel Iris Xe Graphics"}
		cpus := []string{"Intel Core i7-10700K", "AMD Ryzen 5 3600", "Intel Core i5-10400"}
		props.GPU = gpus[rand.Intn(len(gpus))]
		props.CPU = cpus[rand.Intn(len(cpus))]
		props.HardwareConcurrency = 2 + rand.Intn(8)
		props.DeviceMemory = 4 + rand.Intn(12)

		charge := []bool{true, false}
		props.IsCharging = charge[rand.Intn(len(charge))]
		props.BatteryVolume = 0.5 + rand.Float64()/2
		//props.LocalIP = ""
		//props.PublicIP = "" // GetProxy()
	} else {
		props.DeviceProfile.UserAgent = *account.UserAgent
		props.DeviceProfile.Platform = *account.Platform
		props.DeviceProfile.Screen.Width = *account.ScreenWidth
		props.DeviceProfile.Screen.Height = *account.ScreenHeight

		props.RandomFrequency = *account.Frequency
		props.RandomStart = *account.Start
		props.RandomStop = *account.Stop

		props.BufferSize = *account.BufferSize
		props.InputChannels = *account.InputChannels
		props.OutputChannels = *account.OutputChannels

		props.GPU = *account.GPU
		props.CPU = *account.CPU
		props.HardwareConcurrency = *account.HardwareConcurrency
		props.DeviceMemory = *account.DeviceMemory

		props.IsCharging = *account.IsCharging
		props.BatteryVolume = *account.BatteryVolume

		props.LocalIP = *account.LocalIP
		props.PublicIP = *account.PublicIP
	}

	return props
}

// Удаление директории сессии
func clearSessionData(accountID int64) error {
	sessionPath := filepath.Join("./chrome-profile", strconv.Itoa(int(accountID)))
	if err := os.RemoveAll(sessionPath); err != nil {
		return fmt.Errorf("failed to remove session directory: %w", err)
	}
	return nil
}

func injectNavigatorProperties(ctx context.Context, deviceProfile schem.DeviceProfile, cpu string, hardwareConcurrency int, deviceMemory int) error {
	jsScripts := []string{
		// Установка User-Agent
		fmt.Sprintf(`Object.defineProperty(navigator, 'userAgent', { get: () => '%s' });`, deviceProfile.UserAgent),

		// Изменение платформы
		fmt.Sprintf(`Object.defineProperty(navigator, 'platform', { get: () => '%s' });`, deviceProfile.Platform),

		// Изменение аппаратных параметров
		fmt.Sprintf(`Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => %d });`, hardwareConcurrency),
		fmt.Sprintf(`Object.defineProperty(navigator, 'deviceMemory', { get: () => %d });`, deviceMemory),
		fmt.Sprintf(`Object.defineProperty(navigator, 'cpu', { get: () => '%s' });`, cpu),
	}

	// Выполнение JS-скриптов для navigator
	for _, script := range jsScripts {
		log.Printf("Injecting navigator script: %s", script)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting navigator script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

func injectCanvasAndWebGL(ctx context.Context, gpu string, cpu string) error {
	jsScripts := []string{
		// Подделка Canvas
		fmt.Sprintf(`HTMLCanvasElement.prototype.toDataURL = (() => 'data:image/png;base64,data-%s');`, cpu+gpu),

		// Подделка WebGL Renderer
		fmt.Sprintf(`
			const getExtension = WebGLRenderingContext.prototype.getExtension;
			WebGLRenderingContext.prototype.getExtension = function(name) {
				if (name === 'WEBGL_debug_renderer_info') {
					return {
						UNMASKED_VENDOR_WEBGL: "Custom Vendor",
						UNMASKED_RENDERER_WEBGL: "%s (%s)"
					};
				}
				return getExtension.call(this, name);
			};
		`, gpu, cpu),
	}

	// Выполнение JS-скриптов для Canvas и WebGL
	for _, script := range jsScripts {
		log.Printf("Injecting canvas/webgl script: %s", script)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting canvas/webgl script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

func injectWebRTCProperties(ctx context.Context, localIP string, publicIP string) error {
	jsScript := []string{
		fmt.Sprintf(`
		(() => {
			const originalCreateOffer = RTCPeerConnection.prototype.createOffer;
			RTCPeerConnection.prototype.createOffer = function (...args) {
				return originalCreateOffer.apply(this, args).then(offer => {
					offer.sdp = offer.sdp.replace(/a=candidate:.*\r\n/g, match => {
						if (match.includes("typ host")) {
							return match.replace(/[\d.]+ \d+ typ host/, "%s 9 typ host");
						}
						if (match.includes("typ srflx")) {
							return match.replace(/[\d.]+ \d+ typ srflx/, "%s 9 typ srflx");
						}
						return match;
					});
					return offer;
				});
			};
		})();
		`, localIP, publicIP),
	}

	// Выполнение JS-скриптов для WebRTC
	for _, script := range jsScript {
		log.Printf("Injecting webrtc script: %s", script)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting webrtc script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

func injectScreenAndAudioProperties(
	ctx context.Context,
	deviceProfile schem.DeviceProfile,
	bufferSize int,
	inputChannels int,
	outputChannels int,
	randomFrequency int,
	randomStart string,
	randomStop string,
) error {

	jsScripts := []string{
		// Setting screen properties
		fmt.Sprintf(`
       window.customScreen = { 
           width: %d, 
           height: %d 
       };
       `, deviceProfile.Screen.Width, deviceProfile.Screen.Height),

		// Generating audio fingerprint
		fmt.Sprintf(`
			window.audioFingerprint = null;
			
			const getAudioFingerprint = async () => {
				const ctx = new (window.AudioContext || window.webkitAudioContext)();
				const oscillator = ctx.createOscillator();
				const analyser = ctx.createAnalyser();
				const gain = ctx.createGain();
				const scriptProcessor = ctx.createScriptProcessor(%d, %d, %d);
			
				oscillator.type = 'triangle';
				oscillator.frequency.value = %d;
				gain.gain.value = 0;
			
				oscillator.connect(analyser);
				analyser.connect(scriptProcessor);
				scriptProcessor.connect(ctx.destination);
			
				oscillator.start(%s);
				oscillator.stop(%s);
			
				return new Promise((resolve, reject) => {
					scriptProcessor.onaudioprocess = () => {
						const freqData = new Uint8Array(analyser.frequencyBinCount);
						analyser.getByteFrequencyData(freqData);
						oscillator.stop();
						scriptProcessor.disconnect();
			
						resolve(freqData.join(","));
					};
			
					oscillator.onerror = (error) => {
						reject("Audio context error: " + error);
					};
				});
			};
			
			getAudioFingerprint().then(fingerprint => {
				window.audioFingerprint = fingerprint;
			}).catch(error => {
				console.error("Error generating audio fingerprint:", error);
			});
			`, bufferSize, inputChannels, outputChannels, randomFrequency, randomStart, randomStop),
	}

	// Executing JavaScript code to set screen properties and generate audio fingerprint
	for _, script := range jsScripts {
		log.Printf("Injecting script: %s", script)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

func injectMediaAndBatteryProperties(ctx context.Context, accountID int64, isCharging bool, batteryVolume float64) error {
	jsScripts := []string{
		// Медиа устройства
		fmt.Sprintf(`
			Object.defineProperty(navigator, 'mediaDevices', {
				get: () => ({
					enumerateDevices: () => Promise.resolve([
						{ kind: 'videoinput', label: 'Integrated Camera', deviceId: '%d-camera' },
						{ kind: 'audioinput', label: 'Built-in Microphone', deviceId: '%d-microphone' }
					])
				})
			});
		`, accountID, accountID),

		// Battery API
		fmt.Sprintf(`
			Object.defineProperty(navigator, 'getBattery', {
				get: () => () => Promise.resolve({
					charging: %v,
					chargingTime: 0,
					dischargingTime: Infinity,
					level: %.2f
				})
			});
		`, isCharging, batteryVolume),
	}

	// Выполнение JS-скриптов для медиа и батареи
	for _, script := range jsScripts {
		log.Printf("Injecting media/battery script: %s", script)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting media/battery script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

// Настройка хрома JS скриптами

func InjectJSProperties(c context.Context, account models.Account) (schem.AccountProperties, error) {

	props := createAccountProperties(account)

	if account.IsAuthenticated != true {
		// Удаление сессии
		if err := clearSessionData(int64(account.ID)); err != nil {
			return schem.AccountProperties{}, fmt.Errorf("error clearing session data: %w", err)
		}
	}

	// Выполняем инъекцию свойств по частям
	if err := injectNavigatorProperties(c, props.DeviceProfile, props.CPU, props.HardwareConcurrency, props.DeviceMemory); err != nil {
		return props, err
	}
	if err := injectCanvasAndWebGL(c, props.GPU, props.CPU); err != nil {
		return props, err
	}
	//if err := injectWebRTCProperties(c, props.localIP, props.publicIP); err != nil {
	//	return err
	//}
	if err := injectScreenAndAudioProperties(
		c,
		props.DeviceProfile,
		props.BufferSize,
		props.InputChannels,
		props.OutputChannels,
		props.RandomFrequency,
		props.RandomStart,
		props.RandomStop,
	); err != nil {
		return props, err
	}
	if err := injectMediaAndBatteryProperties(c, int64(account.ID), props.IsCharging, props.BatteryVolume); err != nil {
		return props, err
	}

	return props, nil
}

// Установка куки

func setCookies(c *gin.Context, account models.Account) error {

	// Десериализация cookies из строки
	var cookies []network.Cookie
	err := json.Unmarshal([]byte(*account.SessionCookies), &cookies)
	if err != nil {
		return fmt.Errorf("failed to parse session cookies: %w", err)
	}

	// Установка cookies через chromedp
	err = chromedp.Run(c, chromedp.ActionFunc(func(ctx context.Context) error {
		for _, cookie := range cookies {
			var expires *cdp.TimeSinceEpoch
			if cookie.Expires > 0 {
				expiresTime := time.Unix(int64(cookie.Expires), 0)
				exp := cdp.TimeSinceEpoch(expiresTime)
				expires = &exp
			}

			// Установка куки
			err := network.SetCookie(cookie.Name, cookie.Value).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithExpires(expires).
				WithHTTPOnly(cookie.HTTPOnly).
				WithSecure(cookie.Secure).
				Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to set cookie %s: %w", cookie.Name, err)
			}
		}
		return nil
	}))

	if err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}

	return nil
}

// Получение куки

func GetSessionCookies(c *gin.Context) (*string, error) {
	// Получаем все куки из запроса
	cookieHeaders := c.Request.Cookies()
	if len(cookieHeaders) == 0 {
		return nil, fmt.Errorf("no cookies found in the request")
	}

	// Формируем строку из всех куки
	var cookies string
	for _, cookie := range cookieHeaders {
		cookies += fmt.Sprintf("%s=%s; ", cookie.Name, cookie.Value)
	}

	// Убираем последний лишний "; " (если куки есть)
	if len(cookies) > 2 {
		cookies = cookies[:len(cookies)-2]
	}

	return &cookies, nil
}

// Настройка Хром драйвера

func SetupChromeDriver(c *gin.Context, account models.Account, cfg config.Config) (cu.Config, error) {

	conf := cu.NewConfig(
		cu.WithContext(c),
	)

	// Настройки ChromeFlags и других параметров
	conf.ChromeFlags = append(conf.ChromeFlags,
		chromedp.Flag("user-data-dir", "./chrome-profile/"+strconv.Itoa(int(account.ID))),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-features", "FontEnumeration"),
		chromedp.Flag("disable-sync", true),                           // Отключение синхронизации Google
		chromedp.Flag("metrics-recording-only", true),                 // Отключение некоторых аналитических данных
		chromedp.Flag("disable-background-timer-throttling", true),    // Отключение замедления таймеров в фоновом режиме
		chromedp.Flag("disable-backgrounding-occluded-windows", true), // Запуск фоновых окон без ограничений
	)
	if cfg.GinMode != "debug" {
		conf.ChromeFlags = append(conf.ChromeFlags,
			chromedp.Flag("disable-extensions", true))
	}

	if account.UserAgent != nil && *account.UserAgent != "" {
		conf.ChromeFlags = append(conf.ChromeFlags,
			chromedp.Flag("headless", true),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
		)
	}
	if account.PublicIP != nil && *account.PublicIP != "" {
		conf.ChromeFlags = append(conf.ChromeFlags,
			chromedp.Flag("proxy-server", account.PublicIP), // Настройка HTTP-прокси

		)
	}

	if account.SessionCookies != nil && *account.SessionCookies != "" {
		if err := setCookies(c, account); err != nil {
			log.Printf("Error set cookies: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set cookies"})
			return cu.Config{}, err
		}

	}

	return conf, nil

}
