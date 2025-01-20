package utils

import (
	schem "bank_parser_backend_go/internal/schemas"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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
		`HTMLCanvasElement.prototype.toDataURL = (() => 'data:image/png;base64,fakedata');`,

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
	jsScripts := []string{
		// Установка свойств экрана
		fmt.Sprintf(`
		window.customScreen = { 
			width: %d, 
			height: %d 
		};
		`, deviceProfile.Screen.Width, deviceProfile.Screen.Height),

		// Создание простого аудио fingerprint
		fmt.Sprintf(`
			const context = new (window.AudioContext || window.webkitAudioContext)();
			const oscillator = context.createOscillator();
			oscillator.type = 'triangle';
			oscillator.frequency.value = %d;
			oscillator.connect(context.destination);
			oscillator.start(%s);
			oscillator.stop(%s);
		`, randomFrequency, randomStart, randomStop),
	}

	// Выполнение JS-скриптов для экрана и аудио
	for _, script := range jsScripts {
		log.Printf("Injecting screen/audio script: %s", script)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, nil)); err != nil {
			log.Printf("Error injecting screen/audio script: %v", err)
			return fmt.Errorf("JS injection failed: %w", err)
		}
	}

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
	randomStop := fmt.Sprintf("%.2f", 0.01+rand.Float64()/10)

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
