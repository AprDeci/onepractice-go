package service

import (
	"log/slog"

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

	// TODO: wire real SMTP sender after env verified. QQ SMTP over 465 needs implicit TLS.
	return nil
}
