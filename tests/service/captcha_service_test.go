package service_test

import (
	"context"
	"errors"
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
