package utils

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	schem "bank_parser_backend_go/internal/schemas"
	"context"
	"fmt"
	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:91.0) Gecko/20100101 Firefox/91.0",
			Platform:  "Win64",
			Screen:    struct{ Width, Height int }{2560, 1440},
		},
		{
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_5_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36",
			Platform:  "MacIntel",
			Screen:    struct{ Width, Height int }{2880, 1800},
		},
		{
			UserAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:90.0) Gecko/20100101 Firefox/90.0",
			Platform:  "Linux x86_64",
			Screen:    struct{ Width, Height int }{3440, 1440},
		},
		{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
			Platform:  "Win32",
			Screen:    struct{ Width, Height int }{1366, 768},
		},
		{
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36",
			Platform:  "MacIntel",
			Screen:    struct{ Width, Height int }{1280, 800},
		},
		{
			UserAgent: "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0",
			Platform:  "Linux x86_64",
			Screen:    struct{ Width, Height int }{1600, 900},
		},
		{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
			Platform:  "Win32",
			Screen:    struct{ Width, Height int }{1680, 1050},
		},
		{
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36",
			Platform:  "MacIntel",
			Screen:    struct{ Width, Height int }{1920, 1200},
		},
		{
			UserAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0",
			Platform:  "Linux x86_64",
			Screen:    struct{ Width, Height int }{1280, 720},
		},
		{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36 OPR/77.0.4054.277",
			Platform:  "Win32",
			Screen:    struct{ Width, Height int }{2560, 1600},
		},
		{
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36",
			Platform:  "MacIntel",
			Screen:    struct{ Width, Height int }{3840, 2160},
		},
	}

	return devices[rand.Intn(len(devices))]
}

func createAccountProperties(account models.Account) schem.AccountProperties {
	var props schem.AccountProperties

	props.PublicIP = *account.PublicIP

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
	}

	return props
}

// Удаление директории сессии

func ClearSessionDir(accountID int64, ignoreNotExist bool) (string, error) {
	sessionPath := filepath.Join("./chrome-profile", strconv.Itoa(int(accountID)))

	// Проверяем, существует ли директория
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		if ignoreNotExist {
			// Возвращаем успех, если нужно игнорировать ошибку отсутствия директории
			return "Directory does not exist, but ignored", nil
		}
		// Возвращаем ошибку, если игнорировать её не нужно
		return "", fmt.Errorf("directory does not exist: %s", sessionPath)
	}

	// Удаляем директорию
	if err := os.RemoveAll(sessionPath); err != nil {
		return "", fmt.Errorf("failed to remove session directory: %w", err)
	}
	return "Directory deleted successfully", nil
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
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting canvas/webgl script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

func injectWebRTCProperties(ctx context.Context, publicIP string) error {
	jsScript := fmt.Sprintf(`
    (() => {
        const originalCreateOffer = RTCPeerConnection.prototype.createOffer;
        RTCPeerConnection.prototype.createOffer = function (...args) {
            return originalCreateOffer.apply(this, args).then(offer => {
                // Заменяем все локальные IP-адреса на publicIP
                offer.sdp = offer.sdp.replace(/a=candidate:[^\n]+\s+[\d.]+\s+\d+\s+typ\s+(host|srflx)/g, (match, type) => {
                    return match.replace(/[\d.]+\s+\d+\s+typ\s+(host|srflx)/, "%s 9 typ " + type);
                });
                return offer;
            });
        };
    })();
    `, publicIP)

	// Выполнение JS-скрипта для WebRTC
	if err := chromedp.Run(ctx, chromedp.Evaluate(jsScript, nil)); err != nil {
		log.Printf("Error injecting webrtc script: %v", err)
		return fmt.Errorf("JS injection failed: %w", err)
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
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting media/battery script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

	return nil
}

// Настройка хрома JS скриптами

func InjectJSProperties(ctx context.Context, account models.Account) (schem.AccountProperties, error) {

	props := createAccountProperties(account)

	// Выполняем инъекцию свойств по частям
	if err := injectNavigatorProperties(ctx, props.DeviceProfile, props.CPU, props.HardwareConcurrency, props.DeviceMemory); err != nil {
		return props, fmt.Errorf("failed to enable network tracking: %w", err)
	}
	if err := injectCanvasAndWebGL(ctx, props.GPU, props.CPU); err != nil {
		return props, err
	}
	if err := injectWebRTCProperties(ctx, props.PublicIP); err != nil {
		return props, err
	}
	if err := injectScreenAndAudioProperties(
		ctx,
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
	if err := injectMediaAndBatteryProperties(ctx, int64(account.ID), props.IsCharging, props.BatteryVolume); err != nil {
		return props, err
	}

	return props, nil
}

// Парсинг cookies из строки в []network.CookieParam

func parseCookieString(cookieString string, cfg config.Config) ([]network.CookieParam, error) {
	parts := strings.Split(cookieString, "; ")
	var cookies []network.CookieParam

	for _, part := range parts {
		// Разделяем имя и значение
		cookieParts := strings.SplitN(part, "=", 2)
		if len(cookieParts) != 2 {
			return nil, fmt.Errorf("invalid cookie format: %s", part)
		}

		expiresAt := time.Now().Add(31536000 * time.Second) // 1 год
		expiresTime := cdp.TimeSinceEpoch(expiresAt)

		cookies = append(cookies, network.CookieParam{
			Name:     cookieParts[0],
			Value:    cookieParts[1],
			Domain:   cfg.AlphaUrl,
			Path:     "/",
			HTTPOnly: false,
			Secure:   false,
			Expires:  &expiresTime,
		})
	}

	return cookies, nil
}

// Установка куки

func setCookies(ctx context.Context, account models.Account, cfg config.Config) error {
	cookies, err := parseCookieString(*account.SessionCookies, cfg)
	if err != nil {
		return fmt.Errorf("failed to parse session cookies: %w", err)
	}

	// Установка cookies через chromedp
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		for _, cookie := range cookies {
			err := network.SetCookie(cookie.Name, cookie.Value).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithHTTPOnly(cookie.HTTPOnly).
				WithSecure(cookie.Secure).
				WithExpires(cookie.Expires).
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

func GetSessionCookies(ctx context.Context) (string, error) {
	// Получаем cookies через network.GetCookies
	var cookies []*network.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))
	if err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	// Формируем строку из всех cookies
	var cookieString string
	for _, cookie := range cookies {
		cookieString += fmt.Sprintf("%s=%s; ", cookie.Name, cookie.Value)
	}

	// Убираем последний лишний "; "
	if len(cookieString) > 2 {
		cookieString = cookieString[:len(cookieString)-2]
	}

	return cookieString, nil
}

// Настройка Хром драйвера

func SetupChromeDriver(ctx context.Context, account models.Account, cfg config.Config) (context.Context, context.CancelFunc, error) {
	// Создаем конфигурацию Chrome
	conf := cu.NewConfig(
		cu.WithContext(ctx),
	)

	// Настройки ChromeFlags
	conf.ChromeFlags = append(conf.ChromeFlags,
		chromedp.Flag("user-data-dir", "./chrome-profile/"+strconv.Itoa(int(account.ID))),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-features", "FontEnumeration"),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("proxy-server", *account.PublicIP),
	)
	if cfg.GinMode != "debug" {
		conf.ChromeFlags = append(conf.ChromeFlags,
			chromedp.Flag("disable-extensions", true),
			chromedp.Flag("headless", true),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
		)
	}

	if account.UserAgent != nil && *account.UserAgent != "" {
		conf.ChromeFlags = append(conf.ChromeFlags,
			chromedp.Flag("user-agent", *account.UserAgent),
		)
	}

	if account.IsAuthenticated != true {
		// Удаление сессии
		if _, err := ClearSessionDir(int64(account.ID), true); err != nil {
			return nil, nil, fmt.Errorf("error clearing session data: %w", err)
		}
	}

	// Создание нового контекста Chrome через cu.New
	ctx, cancel, err := cu.New(conf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize ChromeDriver context: %w", err)
	}

	// Устанавливаем cookies в контекст браузера
	if account.SessionCookies != nil && *account.SessionCookies != "" {
		if err := setCookies(ctx, account, cfg); err != nil {
			log.Printf("Error setting cookies: %v", err)
			cancel() // Закрываем контекст при ошибке
			return nil, nil, fmt.Errorf("failed to set cookies: %w", err)
		}
	}

	return ctx, cancel, nil
}
