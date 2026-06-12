package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"onepractice-golang/internal/config"
	"onepractice-golang/internal/model"
	"onepractice-golang/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CaptchaPurposeRegister      = "register"
	CaptchaPurposeResetPassword = "reset_password"
)

type CaptchaService struct {
	db     *gorm.DB
	mailer *MailService
	redis  *redis.Client
}

func NewCaptchaService(db *gorm.DB, mailCfg config.MailConfig, redisClient *redis.Client) *CaptchaService {
	return &CaptchaService{db: db, mailer: NewMailService(mailCfg), redis: redisClient}
}

func (s *CaptchaService) SendEmailCaptcha(email string) error {
	if s.db == nil && s.redis == nil {
		return ErrDatabaseDisabled
	}

	email = normalizeEmail(email)
	if email == "" {
		return ErrInvalidParam
	}

	if err := s.checkSendCooldown(email); err != nil {
		return err
	}

	code, err := randomCode()
	if err != nil {
		return err
	}

	if err := s.storeCode(email, code, CaptchaPurposeRegister); err != nil {
		return err
	}

	return s.mailer.Send(email, "onepractice 验证码", fmt.Sprintf("您的验证码是：%s，5分钟内有效。", code))
}

func (s *CaptchaService) VerifyRegister(email, code string) error {
	return s.consumeCode(normalizeEmail(email), code, CaptchaPurposeRegister)
}

func (s *CaptchaService) VerifyResetPassword(email, code string) (string, error) {
	if s.db == nil {
		return "", ErrDatabaseDisabled
	}

	email = normalizeEmail(email)
	if exists, err := s.emailExists(email); err != nil {
		return "", err
	} else if !exists {
		return "", ErrInvalidParam
	}

	if err := s.consumeCode(email, code, CaptchaPurposeRegister); err != nil {
		return "", err
	}

	token := strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := s.storeResetToken(email, token); err != nil {
		return "", err
	}
	return token, nil
}

func (s *CaptchaService) ConsumeResetToken(email, token string) error {
	if s.redis == nil {
		return nil
	}
	if email == "" || token == "" {
		return ErrInvalidParam
	}

	ctx := context.Background()
	key := resetTokenKey(email, token)
	storedEmail, err := s.redis.Get(ctx, key).Result()
	if err != nil || storedEmail != email {
		return ErrInvalidParam
	}
	return s.redis.Del(ctx, key).Err()
}

func (s *CaptchaService) UsesRedis() bool {
	return s.redis != nil
}

func (s *CaptchaService) consumeCode(email, code, purpose string) error {
	if s.db == nil && s.redis == nil {
		return ErrDatabaseDisabled
	}
	if email == "" || code == "" {
		return ErrInvalidParam
	}
	if s.redis != nil {
		ctx := context.Background()
		key := captchaCodeKey(email, purpose)
		storedCode, err := s.redis.Get(ctx, key).Result()
		if err != nil || storedCode != code {
			return ErrCaptchaInvalid
		}
		return s.redis.Del(ctx, key).Err()
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var record model.EmailCode
		err := tx.Where("email = ? and purpose = ? and consumed_at is null", email, purpose).
			Order("created_at desc").
			First(&record).Error
		if err != nil {
			return ErrCaptchaInvalid
		}
		if time.Now().After(record.ExpiresAt) || record.Code != code {
			return ErrCaptchaInvalid
		}
		now := time.Now()
		return tx.Model(&record).Update("consumed_at", &now).Error
	})
}

func (s *CaptchaService) checkSendCooldown(email string) error {
	if s.redis != nil {
		ctx := context.Background()
		ok, err := s.redis.SetNX(ctx, captchaCooldownKey(email), "1", 60*time.Second).Result()
		if err != nil {
			return err
		}
		if !ok {
			return ErrEmailSendWait
		}
		return nil
	}

	var recent int64
	if err := s.db.Model(&model.EmailCode{}).
		Where("email = ? and created_at > ?", email, time.Now().Add(-60*time.Second)).
		Count(&recent).Error; err != nil {
		return err
	}
	if recent > 0 {
		return ErrEmailSendWait
	}
	return nil
}

func (s *CaptchaService) storeCode(email, code, purpose string) error {
	if s.redis != nil {
		return s.redis.Set(context.Background(), captchaCodeKey(email, purpose), code, 5*time.Minute).Err()
	}

	record := model.EmailCode{
		Email:     email,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	return s.db.Create(&record).Error
}

func (s *CaptchaService) storeResetToken(email, token string) error {
	if s.redis != nil {
		return s.redis.Set(context.Background(), resetTokenKey(email, token), email, 5*time.Minute).Err()
	}

	record := model.PasswordResetToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	return s.db.Create(&record).Error
}

func captchaCodeKey(email, purpose string) string {
	return "onepractice:captcha:email:" + purpose + ":" + email
}

func captchaCooldownKey(email string) string {
	return "onepractice:captcha:email:cooldown:" + email
}

func resetTokenKey(email, token string) string {
	return "onepractice:password-reset:" + email + ":" + token
}

func (s *CaptchaService) emailExists(email string) (bool, error) {
	encEmail, err := utils.LegacyAESEncrypt(email)
	if err != nil {
		return false, err
	}
	var count int64
	if err := s.db.Model(&model.User{}).Where("email in ?", []string{email, encEmail}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
