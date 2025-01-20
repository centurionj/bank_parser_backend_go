package controllers

import (
	"bank_parser_backend_go/internal/config"
	"bank_parser_backend_go/internal/models"
	schem "bank_parser_backend_go/internal/schemas"
	"context"
	"errors"
	"fmt"
	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type AccountController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

// Конструктор контроллеров Account

func NewAccountController(db *gorm.DB, cfg *config.Config) *AccountController {
	return &AccountController{DB: db, Cfg: cfg}
}

// Метод получения Account по его id

func (ac *AccountController) GetAccount(c *gin.Context) (*models.Account, error) {
	var accountIDRequest schem.AccountIDRequest
	if err := c.ShouldBindJSON(&accountIDRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id_request format"})
		return nil, err
	}

	var account models.Account
	if err := ac.DB.First(&account, accountIDRequest.AccountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return nil, err
	}

	return &account, nil
}

func (ac *AccountController) DelAccountProfileDir(c *gin.Context) error {
	var request schem.AccountAccountProfileDirRequest

	// Parse the JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return fmt.Errorf("invalid request format: %w", err)
	}

	// Define the base directory where the folders are located
	baseDir := "./chrome-profile/"

	// Iterate through the account IDs and attempt to delete the directories
	for _, accountID := range request.AccountIDs {
		dirPath := filepath.Join(baseDir, fmt.Sprintf("%d", accountID))

		// Attempt to remove the directory
		if err := os.RemoveAll(dirPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete directory for account ID %d: %s", accountID, err.Error())})
			return fmt.Errorf("deletion ERROR: %w", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Directories deleted successfully"})

	return nil
}

func (ac *AccountController) DelAccountProfileDirHandler(c *gin.Context) {
	result := ac.DelAccountProfileDir(c)

	c.JSON(http.StatusOK, &result)
}

func getRandomDeviceProfile() schem.DeviceProfile {
	// Набор профилей для различных устройств
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

// Удаление директории сессии
func clearSessionData(accountID int64) error {
	sessionPath := filepath.Join("./chrome-profile", strconv.Itoa(int(accountID)))
	if err := os.RemoveAll(sessionPath); err != nil {
		return fmt.Errorf("failed to remove session directory: %w", err)
	}
	return nil
}

func injectNavigatorProperties(ctx context.Context, deviceProfile schem.DeviceProfile, cpu string) error {
	jsScripts := []string{
		// Установка User-Agent
		fmt.Sprintf(`Object.defineProperty(navigator, 'userAgent', { get: () => '%s' });`, deviceProfile.UserAgent),

		// Изменение платформы
		fmt.Sprintf(`Object.defineProperty(navigator, 'platform', { get: () => '%s' });`, deviceProfile.Platform),

		// Изменение аппаратных параметров
		fmt.Sprintf(`Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => %d });`, 2+rand.Intn(8)),
		fmt.Sprintf(`Object.defineProperty(navigator, 'deviceMemory', { get: () => %d });`, 4+rand.Intn(12)),
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

func injectScreenAndAudioProperties(ctx context.Context, deviceProfile schem.DeviceProfile, randomFrequency int, randomStart string, randomStop string) error {
	bufferSizes := []int{256, 512, 1024, 2048, 4096, 8192, 16384}
	bufferSize := bufferSizes[rand.Intn(len(bufferSizes))]
	inputChannels := rand.Intn(2) + 1
	outputChannels := rand.Intn(2) + 1

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

	//var fingerprint string
	//var retries int
	//const maxRetries = 30                 // Увеличил количество попыток
	//const retryInterval = 1 * time.Second // Интервал между попытками
	//
	//for retries = 0; retries < maxRetries; retries++ {
	//	err := chromedp.Run(ctx,
	//		chromedp.Evaluate(`window.audioFingerprint || "Fingerprint not available yet"`, &fingerprint),
	//	)
	//	if err != nil {
	//		return fmt.Errorf("error retrieving fingerprint: %w", err)
	//	}
	//
	//	if fingerprint != "Fingerprint not available yet" && fingerprint != "" {
	//		log.Printf("Fingerprint: %s", fingerprint)
	//		return nil
	//	}
	//
	//	time.Sleep(retryInterval) // Задержка между попытками
	//}
	//
	//log.Printf("Fingerprint not available after %d retries", retries)
	//println("CONTEXT", ctx)
	//return fmt.Errorf("Fingerprint not available yet")

	return nil
}

func injectMediaAndBatteryProperties(ctx context.Context, accountID int64) error {
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
					charging: true,
					chargingTime: 0,
					dischargingTime: Infinity,
					level: %.2f
				})
			});
		`, 0.5+rand.Float64()/2),
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

func injectJSProperties(ctx context.Context, accountID int64) error {
	// Получаем случайный профиль устройства
	deviceProfile := getRandomDeviceProfile()

	// Рандомизация параметров
	randomFrequency := 8000 + rand.Intn(4000)
	randomStart := fmt.Sprintf("%.2f", rand.Float64()/10)
	randomStop := fmt.Sprintf("%.1f", 0.1+rand.Float64()/10)

	// Рандомизация оборудования
	gpus := []string{"NVIDIA GeForce GTX 1660", "AMD Radeon RX 580", "Intel Iris Xe Graphics"}
	cpus := []string{"Intel Core i7-10700K", "AMD Ryzen 5 3600", "Intel Core i5-10400"}
	gpu := gpus[rand.Intn(len(gpus))]
	cpu := cpus[rand.Intn(len(cpus))]

	//localIP := "http://localhost:8080"
	//publicIP := "http://publichost:8080"

	// Удаление предыдущей сессии
	if err := clearSessionData(accountID); err != nil {
		return fmt.Errorf("error clearing session data: %w", err)
	}

	// Выполняем инъекцию свойств по частям
	if err := injectNavigatorProperties(ctx, deviceProfile, cpu); err != nil {
		return err
	}
	if err := injectCanvasAndWebGL(ctx, gpu, cpu); err != nil {
		return err
	}
	//if err := injectWebRTCProperties(ctx, localIP, publicIP); err != nil {
	//	return err
	//}
	if err := injectScreenAndAudioProperties(ctx, deviceProfile, randomFrequency, randomStart, randomStop); err != nil {
		return err
	}
	if err := injectMediaAndBatteryProperties(ctx, accountID); err != nil {
		return err
	}

	return nil
}

func (ac *AccountController) getUserAgent(ctx context.Context) (string, error) {
	var userAgent string
	err := chromedp.Run(ctx, chromedp.Evaluate(`navigator.userAgent`, &userAgent))
	if err != nil {
		return "", fmt.Errorf("failed to retrieve User-Agent: %w", err)
	}
	return userAgent, nil
}

func (ac *AccountController) getSessionCookies(ctx context.Context) (*string, error) {
	var cookies string

	// Выполнение JS для извлечения cookies
	err := chromedp.Run(ctx, chromedp.Evaluate(`document.cookie`, &cookies))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cookies using JS: %w", err)
	}

	return &cookies, nil
}

func injectDeviceProperties(ctx context.Context, account *models.Account, ac *AccountController) error {
	// Извлекаем данные User-Agent
	userAgent, err := ac.getUserAgent(ctx) // Вызов на экземпляре ac
	if err != nil {
		return fmt.Errorf("failed to get User-Agent: %w", err)
	}

	// Извлекаем Session Cookies
	sessionCookies, err := ac.getSessionCookies(ctx) // То же самое здесь
	if err != nil {
		return fmt.Errorf("failed to get session cookies: %w", err)
	}

	// Извлекаем данные Navigator
	var platform string
	err = chromedp.Run(ctx, chromedp.Evaluate(`navigator.platform`, &platform))
	if err != nil {
		return fmt.Errorf("failed to get navigator platform: %w", err)
	}

	var hardwareConcurrency int
	err = chromedp.Run(ctx, chromedp.Evaluate(`navigator.hardwareConcurrency`, &hardwareConcurrency))
	if err != nil {
		return fmt.Errorf("failed to get hardwareConcurrency: %w", err)
	}

	var deviceMemory int
	err = chromedp.Run(ctx, chromedp.Evaluate(`navigator.deviceMemory`, &deviceMemory))
	if err != nil {
		return fmt.Errorf("failed to get deviceMemory: %w", err)
	}

	// Извлечение данных экрана
	var screenWidth, screenHeight int
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		(() => {
			if (window.customScreen) {
				return { width: window.customScreen.width, height: window.customScreen.height };
			} else {
				return { width: window.innerWidth, height: window.innerHeight };
			}
		})()
	`, &struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}{
		screenWidth,
		screenHeight,
	}))
	if err != nil {
		return fmt.Errorf("failed to get custom or default screen size: %w", err)
	}

	// Извлекаем информацию о Canvas
	var canvasFingerprint string
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		var canvas = document.createElement('canvas');
		var context = canvas.getContext('2d');
		context.textBaseline = "top";
		context.font = "14px 'Arial'";
		context.fillText("sample text", 2, 2);
		canvas.toDataURL();
	`, &canvasFingerprint))
	if err != nil {
		return fmt.Errorf("failed to get Canvas fingerprint: %w", err)
	}

	// Извлекаем информацию о WebGL
	var webglVendor, webglRenderer string
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		const gl = document.createElement('canvas').getContext('webgl');
		const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
		const vendor = gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL);
		const renderer = gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL);
		[vendor, renderer]
	`, &[]interface{}{&webglVendor, &webglRenderer}))
	if err != nil {
		return fmt.Errorf("failed to get WebGL info: %w", err)
	}

	// Извлекаем данные WebRTC
	var localIPRaw interface{}
	err = chromedp.Run(ctx, chromedp.Evaluate(`
    new Promise((resolve) => {
        const pc = new RTCPeerConnection({ iceServers: [] });
        pc.onicecandidate = (event) => {
            if (event && event.candidate) {
                const match = event.candidate.candidate.match(/(\d+\.\d+\.\d+\.\d+)/);
                if (match) {
                    resolve(match[1]); // Возвращаем только IP-адрес
                }
            }
        };
        pc.createDataChannel("");
        pc.createOffer().then(offer => pc.setLocalDescription(offer))
        .catch(() => resolve(null)); // Возвращаем null в случае ошибки
    }).catch(() => resolve(null)); // Возвращаем null, если возникает исключение
`, &localIPRaw))
	if err != nil {
		return fmt.Errorf("failed to get WebRTC local IP: %w", err)
	}

	localIP, ok := localIPRaw.(string)
	if !ok || localIP == "" {
		log.Println("WebRTC returned an empty or invalid IP address.")
		localIP = "" // Заглушка
	}

	// Получение информации об аудио и батарее
	//var audioFingerprint string
	//err = chromedp.Run(ctx, chromedp.Evaluate(`
	//	(() => {
	//		if (window.customAudioFingerprint) {
	//			return window.customAudioFingerprint;
	//		}
	//		const ctx = new (window.AudioContext || window.webkitAudioContext)();
	//		const oscillator = ctx.createOscillator();
	//		const analyser = ctx.createAnalyser();
	//		const gain = ctx.createGain();
	//		const scriptProcessor = ctx.createScriptProcessor(4096, 1, 1);
	//
	//		oscillator.type = "triangle";
	//		oscillator.frequency.value = 10000;
	//		gain.gain.value = 0;
	//
	//		oscillator.connect(analyser);
	//		analyser.connect(scriptProcessor);
	//		scriptProcessor.connect(ctx.destination);
	//		oscillator.start(0);
	//
	//		return new Promise(resolve => {
	//			scriptProcessor.onaudioprocess = () => {
	//				const freqData = new Uint8Array(analyser.frequencyBinCount);
	//				analyser.getByteFrequencyData(freqData);
	//				oscillator.stop();
	//				scriptProcessor.disconnect();
	//				resolve(freqData.join(""));
	//			};
	//		});
	//	})();
	//`, &audioFingerprint))
	//if err != nil {
	//	return fmt.Errorf("failed to get audio fingerprint: %w", err)
	//}

	var audioFingerprint interface{}
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		(() => {
			return new Promise(async (resolve) => {
				if (window.customAudioFingerprint) {
					console.log('Returning customAudioFingerprint:', window.customAudioFingerprint);
					resolve(window.customAudioFingerprint);
					return;
				}
	
				try {
					const ctx = new (window.AudioContext || window.webkitAudioContext)();
					const oscillator = ctx.createOscillator();
					const analyser = ctx.createAnalyser();
					const gain = ctx.createGain();
					const scriptProcessor = ctx.createScriptProcessor(4096, 1, 1);
	
					oscillator.type = "triangle";
					oscillator.frequency.value = 10000;
					gain.gain.value = 0;
	
					oscillator.connect(analyser);
					analyser.connect(scriptProcessor);
					scriptProcessor.connect(ctx.destination);
					oscillator.start(0);
	
					scriptProcessor.onaudioprocess = () => {
						const freqData = new Uint8Array(analyser.frequencyBinCount);
						analyser.getByteFrequencyData(freqData);
						oscillator.stop();
						scriptProcessor.disconnect();
						
						// Преобразуем массив в строку через JSON.stringify
						const freqDataStr = JSON.stringify(freqData);
						resolve(freqDataStr);
					};
				} catch (e) {
					resolve(""); // Возвращаем пустую строку при ошибке
				}
			});
		})();
	`, &audioFingerprint))

	log.Printf("Received audio fingerprint: %s", audioFingerprint)
	//if err != nil {
	//	return fmt.Errorf("failed to get audio fingerprint: %w", err)
	//}

	//var batteryLevel float64
	//err = chromedp.Run(ctx, chromedp.Evaluate(`navigator.getBattery().then(battery => battery.level)`, &batteryLevel))
	//if err != nil {
	//	return fmt.Errorf("failed to get battery level: %w", err)
	//}

	// Заполнение структуры Account
	account.UserAgent = &userAgent
	account.SessionCookies = sessionCookies
	account.NavigatorPlatform = &platform
	account.HardwareConcurrency = &hardwareConcurrency
	account.DeviceMemory = &deviceMemory
	account.ScreenWidth = &screenWidth
	account.ScreenHeight = &screenHeight
	account.CanvasFingerprint = &canvasFingerprint
	account.WebglVendor = &webglVendor
	account.WebglRenderer = &webglRenderer
	account.LocalIP = &localIP
	//account.AudioFingerprint = &audioFingerprint
	//account.BatteryLevel = &batteryLevel
	account.AudioFingerprint = nil
	account.BatteryLevel = nil

	// Обновляем запись в базе данных
	return nil
}

func (ac *AccountController) AuthAccount(c *gin.Context) {
	account, err := ac.GetAccount(c)
	if err != nil {
		return
	}

	loginURL := ac.Cfg.AlphaLoginUrl
	if loginURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing AlphaLoginUrl in config"})
		return
	}

	conf := cu.NewConfig(
		cu.WithContext(c),
	)

	// Настройки ChromeFlags и других параметров
	conf.ChromeFlags = append(conf.ChromeFlags,
		chromedp.Flag("user-data-dir", "./chrome-profile/"+strconv.Itoa(int(account.ID))),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-features", "FontEnumeration"),
	)

	ctx, cancel, err := cu.New(conf)
	if err != nil {
		panic(err)
	}
	defer cancel()

	// Внедряем параметры через JS
	if err := injectJSProperties(ctx, int64(account.ID)); err != nil {
		log.Printf("Error injecting JS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to inject browser properties"})
		return
	}

	rand.Seed(time.Now().UnixNano())

	// Выполнение действий в браузере
	err = chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second), // Ждем загрузку страницы

		chromedp.Evaluate(`document.querySelector("button[data-test-id='phone-auth-button']").click()`, nil),
		// Ввод номера телефона
		chromedp.WaitVisible(`input[data-test-id='phoneInput']`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.SendKeys(`input[data-test-id='phoneInput']`, account.PhoneNumber[1:]),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.WaitVisible(`button.phone-auth-browser__submit-button`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.Click(`button.phone-auth-browser__submit-button`, chromedp.NodeVisible),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		// Переход на следующую страницу (проверка, что страница загрузилась)
		chromedp.WaitVisible(`input[data-test-id='card-account-input']`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		// Ввод номера карты
		chromedp.SendKeys(`input[data-test-id='card-account-input']`, account.CardNumber),
		chromedp.Sleep(time.Duration(rand.Intn(3)+1)*time.Second),
		chromedp.Click(`button[data-test-id='card-account-continue-button']`, chromedp.NodeVisible),
		chromedp.Sleep(time.Duration(rand.Intn(5)+1)*time.Second), // Ожидание ввода одноразового кода
	)

	if err != nil {
		log.Printf("Authorization error: %v", err)
		account.IsErrored = true
		ac.DB.Save(&account)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization failed"})
		return
	}

	// Бесконечный цикл ожидания ввода одноразового кода
	otpCode := account.TemporaryCode
	if otpCode == nil || *otpCode == "" {
		println("Waiting for OTP code to appear in the database...")

		startTime := time.Now()
		timeout := 1 * time.Minute // Устанавливаем тайм-аут в 1 минуту

		for {
			// Проверяем, не истек ли тайм-аут
			if time.Since(startTime) > timeout {
				log.Println("Timeout waiting for OTP code")
				account.IsErrored = true
				ac.DB.Save(&account)
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Timeout waiting for OTP code"})
				return
			}

			// Обновляем данные из базы
			if err := ac.DB.First(&account, account.ID).Error; err != nil {
				log.Printf("Database error: %v", err)
				account.IsErrored = true
				ac.DB.Save(&account)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}

			// Проверяем, заполнилось ли поле TemporaryCode
			if account.TemporaryCode != nil && *account.TemporaryCode != "" {
				otpCode = account.TemporaryCode // Обновляем значение
				break
			}

			// Логируем для отладки
			log.Println("OTP code is still empty, retrying...")

			// Пауза перед повторной проверкой
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		}
	}
	time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

	// Ввод одноразового кода по символам
	for index, digit := range *otpCode {
		err = chromedp.Run(ctx,
			chromedp.Click(fmt.Sprintf(`input.code-input__input_fq4wa:nth-of-type(%d)`, index+1), chromedp.NodeVisible),
			chromedp.SendKeys(fmt.Sprintf(`input.code-input__input_fq4wa:nth-of-type(%d)`, index+1), string(digit)),
		)
		if err != nil {
			log.Printf("Error entering OTP digit: %v", err)
			account.IsErrored = true
			ac.DB.Save(&account)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error entering OTP digit"})
			return
		}
		time.Sleep(200 * time.Millisecond) // Небольшая задержка между символами
	}

	// Сбор данных устройства
	//if err := injectDeviceProperties(ctx, account, ac); err != nil {
	//	log.Printf("Error collecting device properties: %v", err)
	//	account.IsErrored = true
	//	ac.DB.Save(account)
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect device properties"})
	//	return
	//}

	account.IsAuthenticated = true
	ac.DB.Save(&account)

	c.JSON(http.StatusOK, gin.H{"message": "Authorization successful"})
}
