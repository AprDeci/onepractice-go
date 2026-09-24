package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"onepractice-golang/internal/service"

	"github.com/redis/go-redis/v9"
)

type fakeMailSender struct{}

func (fakeMailSender) Send(context.Context, string, string, string) error {
	return nil
}

func TestCaptchaServiceSendEmailCaptchaUsesCanceledContext(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1",
	})
	defer redisClient.Close()

	captchaService := service.NewCaptchaServiceWithSender(
		nil,
		redisClient,
		fakeMailSender{},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := captchaService.SendEmailCaptcha(ctx, "user@example.com", service.CaptchaPurposeRegister)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SendEmailCaptcha() error = %v, want context.Canceled", err)
	}

}

func TestCaptchaServiceSendEmailCaptchaUsesExpiredDeadline(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1",
	})
	defer redisClient.Close()

	captchaService := service.NewCaptchaServiceWithSender(
		nil,
		redisClient,
		fakeMailSender{},
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Nanosecond,
	)
	defer cancel()

	time.Sleep(time.Millisecond)

	err := captchaService.SendEmailCaptcha(
		ctx,
		"user@example.com",
		service.CaptchaPurposeRegister,
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"SendEmailCaptcha() error = %v, want context.DeadlineExceeded",
			err,
		)
	}
}

// 验证码必须是 6 位、只含数字与大写字母：不能退回纯数字，也不能混入小写或符号。
func TestRandomCodeUsesUppercaseAlphanumericAlphabet(t *testing.T) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	letters := 0
	for range 500 {
		code, err := service.RandomCode()
		if err != nil {
			t.Fatalf("RandomCode() error = %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("RandomCode() = %q, len = %d, want 6", code, len(code))
		}
		for _, r := range code {
			if !strings.ContainsRune(alphabet, r) {
				t.Fatalf("RandomCode() = %q contains %q outside [0-9A-Z]", code, r)
			}
			if r >= 'A' && r <= 'Z' {
				letters++
			}
		}
	}
	if letters == 0 {
		t.Fatal("RandomCode() produced no letter in 500 draws; sampling is broken")
	}
}
