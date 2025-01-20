package models

import (
	"time"
)

// Модель аккаунта для входа в ЛК

type AccountSetup struct { // В дальнейшем отрефакторить 2 схемы
	// Настройка сессии
	SessionCookies *string `gorm:"type:text;default:null"`

	// Настройки браузера
	UserAgent           *string `gorm:"type:text;default:null"`
	NavigatorPlatform   *string `gorm:"type:varchar(50);default:null"`
	HardwareConcurrency *int    `gorm:"default:null"`
	DeviceMemory        *int    `gorm:"default:null"`

	// Характеристики устройства
	ScreenWidth  *int    `gorm:"default:null"`
	ScreenHeight *int    `gorm:"default:null"`
	GPU          *string `gorm:"type:varchar(100);default:null"`
	CPU          *string `gorm:"type:varchar(100);default:null"`

	// Canvas и WebGL
	CanvasFingerprint *string `gorm:"type:text;default:null"`
	WebglVendor       *string `gorm:"type:varchar(100);default:null"`
	WebglRenderer     *string `gorm:"type:varchar(100);default:null"`

	// WebRTC
	LocalIP  *string `gorm:"type:varchar(45);default:null"`
	PublicIP *string `gorm:"type:varchar(45);default:null"`

	// Аудио fingerprint
	AudioFingerprint *string `gorm:"type:text;default:null"`

	// Заряд батареи
	BatteryLevel *float64 `gorm:"default:null"`
	IsCharging   *bool    `gorm:"default:null"`
}

type Account struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	Title            *string   `gorm:"type:varchar(25);default:null"`
	CardNumber       string    `gorm:"type:char(16);not null"`
	PhoneNumber      string    `gorm:"type:char(11);not null"`
	TemporaryCode    *string   `gorm:"type:char(4);default:null"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	IsActive         bool      `gorm:"default:true"`
	IsAuthenticated  bool      `gorm:"default:false"`
	HasTemporaryCode bool      `gorm:"default:false"`
	IsErrored        bool      `gorm:"default:false"`

	// Настройка сессии
	SessionCookies *string `gorm:"type:text;default:null"`

	// Настройки браузера
	UserAgent           *string `gorm:"type:text;default:null"`
	NavigatorPlatform   *string `gorm:"type:varchar(50);default:null"`
	HardwareConcurrency *int    `gorm:"default:null"`
	DeviceMemory        *int    `gorm:"default:null"`

	// Характеристики устройства
	ScreenWidth  *int    `gorm:"default:null"`
	ScreenHeight *int    `gorm:"default:null"`
	GPU          *string `gorm:"type:varchar(100);default:null"`
	CPU          *string `gorm:"type:varchar(100);default:null"`

	// Canvas и WebGL
	CanvasFingerprint *string `gorm:"type:text;default:null"`
	WebglVendor       *string `gorm:"type:varchar(100);default:null"`
	WebglRenderer     *string `gorm:"type:varchar(100);default:null"`

	// WebRTC
	LocalIP  *string `gorm:"type:varchar(45);default:null"`
	PublicIP *string `gorm:"type:varchar(45);default:null"`

	// Аудио fingerprint
	AudioFingerprint *string `gorm:"type:text;default:null"`

	// Заряд батареи
	BatteryLevel *float64 `gorm:"default:null"`
	IsCharging   *bool    `gorm:"default:null"`
}

// Метод для ренейма таблицы под стиль Django

func (Account) TableName() string {
	return "accounts_account"
}
