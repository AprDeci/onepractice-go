package model

import "time"

type EmailCode struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	Email      string     `gorm:"column:email"`
	Code       string     `gorm:"column:code"`
	Purpose    string     `gorm:"column:purpose"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (EmailCode) TableName() string { return "email_codes" }

type PasswordResetToken struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	Email      string     `gorm:"column:email"`
	Token      string     `gorm:"column:token"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }
