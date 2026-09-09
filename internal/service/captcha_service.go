package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"onepractice-golang/internal/model"
	"onepractice-golang/internal/service/templates"
	"onepractice-golang/internal/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CaptchaPurposeRegister      = "register"
	CaptchaPurposeResetPassword = "reset_password"

	captchaCodeTTL = 5 * time.Minute
)

type CaptchaService struct {
	db     *gorm.DB
	mailer MailSender
	redis  *redis.Client
}

type MailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

func NewCaptchaServiceWithSender(db *gorm.DB, redisClient *redis.Client, mailer MailSender) *CaptchaService {
	return &CaptchaService{db: db, mailer: mailer, redis: redisClient}
}

func (s *CaptchaService) SendEmailCaptcha(ctx context.Context, email, purpose string) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}

	email = normalizeEmail(email)
	purpose = normalizeCaptchaPurpose(purpose)
	if email == "" || purpose == "" {
		return ErrInvalidParam
	}

	if err := s.ensureSendAllowed(ctx, email); err != nil {
		return err
	}

	code, err := randomCode()
	if err != nil {
		return err
	}

	body, err := templates.RenderCaptchaEmail(templates.CaptchaEmailData{
		Code:          code,
		Intro:         captchaEmailIntro(purpose),
		ExpireMinutes: int(captchaCodeTTL / time.Minute),
	})
	if err != nil {
		return err
	}
	if err := s.mailer.Send(ctx, email, "【OnePractice】邮箱验证码", body); err != nil {
		return err
	}

	if err := s.storeCodeAndCooldown(ctx, email, code, purpose); err != nil {
		return err
	}

	return nil
}

func (s *CaptchaService) VerifyRegister(ctx context.Context, email, code string) error {
	return s.consumeCode(ctx, normalizeEmail(email), code, CaptchaPurposeRegister)
}

func (s *CaptchaService) VerifyResetPassword(ctx context.Context, email, code string) (string, error) {
	if s.db == nil {
		return "", ErrDatabaseDisabled
	}
	if s.redis == nil {
		return "", ErrRedisDisabled
	}

	email = normalizeEmail(email)
	purpose := normalizeCaptchaPurpose(CaptchaPurposeResetPassword)
	if exists, err := s.emailExists(ctx, email); err != nil {
		return "", err
	} else if !exists {
		return "", ErrInvalidParam
	}

	if err := s.consumeCode(ctx, email, code, purpose); err != nil {
		return "", err
	}

	token := strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := s.storeResetToken(ctx, email, token); err != nil {
		return "", err
	}
	return token, nil
}

func (s *CaptchaService) ConsumeResetToken(ctx context.Context, email, token string) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}
	if email == "" || token == "" {
		return ErrInvalidParam
	}

	key := resetTokenKey(email, token)
	storedEmail, err := s.redis.Get(ctx, key).Result()
	if err != nil || storedEmail != email {
		return ErrInvalidParam
	}
	return s.redis.Del(ctx, key).Err()
}
func (s *CaptchaService) consumeCode(ctx context.Context, email, code, purpose string) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}
	if email == "" || code == "" {
		return ErrInvalidParam
	}

	key := captchaCodeKey(email, purpose)
	storedCode, err := s.redis.Get(ctx, key).Result()
	if err != nil || storedCode != code {
		return ErrCaptchaInvalid
	}
	return s.redis.Del(ctx, key).Err()
}

func (s *CaptchaService) ensureSendAllowed(ctx context.Context, email string) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}

	exists, err := s.redis.Exists(ctx, captchaCooldownKey(email)).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrEmailSendWait
	}
	return nil
}

func (s *CaptchaService) storeResetToken(ctx context.Context, email, token string) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}
	return s.redis.Set(ctx, resetTokenKey(email, token), email, 5*time.Minute).Err()
}

func (s *CaptchaService) storeCodeAndCooldown(ctx context.Context, email, code, purpose string) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}

	pipe := s.redis.Pipeline()
	pipe.Set(ctx, captchaCodeKey(email, purpose), code, captchaCodeTTL)
	pipe.Set(ctx, captchaCooldownKey(email), "1", 60*time.Second)
	_, err := pipe.Exec(ctx)
	return err
}

func captchaCodeKey(email, purpose string) string {
	return "onepractice:captcha:email:" + purpose + ":" + email
}

func normalizeCaptchaPurpose(purpose string) string {
	switch strings.TrimSpace(purpose) {
	case "", CaptchaPurposeRegister:
		return CaptchaPurposeRegister
	case CaptchaPurposeResetPassword:
		return CaptchaPurposeResetPassword
	default:
		return ""
	}
}

func captchaCooldownKey(email string) string {
	return "onepractice:captcha:email:cooldown:" + email
}

func resetTokenKey(email, token string) string {
	return "onepractice:password-reset:" + email + ":" + token
}

func (s *CaptchaService) emailExists(ctx context.Context, email string) (bool, error) {
	encEmail, err := utils.LegacyAESEncrypt(email)
	if err != nil {
		return false, err
	}
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("email in ?", []string{email, encEmail}).Count(&count).Error; err != nil {
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

func captchaEmailIntro(purpose string) string {
	if purpose == CaptchaPurposeResetPassword {
		return "你正在进行密码重置"
	}
	return "你正在进行注册"
}
