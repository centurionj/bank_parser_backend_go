package models

import (
	"time"
)

// Модель аккаунта для входа в ЛК

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
	SessionCookies   *string   `gorm:"type:text;default:null"`
	UserAgent        *string   `gorm:"type:text;default:null"`
}

// Метод для ренейма таблицы под стиль Django

func (Account) TableName() string {
	return "accounts_account"
}
