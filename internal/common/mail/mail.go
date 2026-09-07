package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"onepractice-golang/internal/common/message_queue"
	"onepractice-golang/internal/config"

	"github.com/redis/go-redis/v9"
	sendflare "github.com/sendflare/sendflare-sdk-go"
)

type Module struct {
	Sender interface {
		Send(ctx context.Context, to, subject, body string) error
	}
	cancel context.CancelFunc
}

type provider struct {
	cfg config.MailConfig
}

type queueSender struct {
	queue *message_queue.Queue
}

type message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewModule(ctx context.Context, cfg config.MailConfig, redisClient *redis.Client) *Module {
	provider := &provider{cfg: cfg}
	if redisClient == nil {
		return &Module{Sender: provider}
	}

	queueCtx, cancel := context.WithCancel(ctx)
	queue := message_queue.NewQueue(
		queueCtx,
		redisClient,
		message_queue.WithTopic("onepractice:mail"),
		message_queue.WithHandler(func(msg message_queue.Message) {
			var item message
			data, err := json.Marshal(msg.Body)
			if err != nil || json.Unmarshal(data, &item) != nil {
				slog.Error("decode mail message", "error", err)
				return
			}
			if err := provider.Send(queueCtx, item.To, item.Subject, item.Body); err != nil {
				slog.Error("send queued mail", "error", err, "to", item.To, "subject", item.Subject)
			}
		}),
	)
	queue.Start()

	return &Module{
		Sender: &queueSender{queue: queue},
		cancel: cancel,
	}
}

func (m *Module) Close() {
	if m != nil && m.cancel != nil {
		m.cancel()
	}
}

func (p *provider) Send(_ context.Context, to, subject, body string) error {
	if p.cfg.Disabled {
		slog.Info("mail disabled, skip send", "to", to, "subject", subject)
		return nil
	}
	if p.cfg.From == "" || p.cfg.APIKey == "" {
		return fmt.Errorf("sendflare config incomplete")
	}

	resp, err := sendflare.NewSendflare(p.cfg.APIKey).SendEmail(sendflare.SendEmailReq{
		From: p.cfg.From, To: to, Subject: subject, Body: body,
	})
	if err != nil {
		return fmt.Errorf("sendflare send mail: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("sendflare send mail failed: %s", resp.Message)
	}
	return nil
}

func (s *queueSender) Send(_ context.Context, to, subject, body string) error {
	if s == nil || s.queue == nil {
		return fmt.Errorf("mail queue is nil")
	}
	_, err := s.queue.Publish(message_queue.NewMessage("", time.Now(), message{
		To: to, Subject: subject, Body: body,
	}))
	return err
}
