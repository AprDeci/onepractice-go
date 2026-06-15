package service

import (
	"fmt"
	"log/slog"

	sendflare "github.com/sendflare/sendflare-sdk-go"

	"onepractice-golang/internal/config"
)

type MailService struct {
	cfg config.MailConfig
}

func NewMailService(cfg config.MailConfig) *MailService {
	return &MailService{cfg: cfg}
}

func (s *MailService) Send(to, subject, body string) error {
	if s.cfg.Disabled {
		slog.Info("mail disabled, skip send", "to", to, "subject", subject, "body", body)
		return nil
	}
	if s.cfg.From == "" || s.cfg.APIKey == "" {
		return fmt.Errorf("sendflare config incomplete")
	}

	client := sendflare.NewSendflare(s.cfg.APIKey)
	resp, err := client.SendEmail(sendflare.SendEmailReq{
		From:    s.cfg.From,
		To:      to,
		Subject: subject,
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("sendflare send mail: %w", err)
	}
	fmt.Println(resp)
	if !resp.Success {
		return fmt.Errorf("sendflare send mail failed: %s", resp.Message)
	}
	return nil
}
