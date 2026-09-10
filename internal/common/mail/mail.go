package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"onepractice-golang/internal/common/message_queue"
	"onepractice-golang/internal/config"

	sendflare "github.com/sendflare/sendflare-sdk-go"
)

// Topic 是邮件发送队列的 topic。
const Topic = "onepractice:mail"

// Sender 负责实际发送邮件。
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Module 持有邮件发送器。
type Module struct {
	Sender Sender
}

type provider struct{ cfg config.MailConfig }

type queueSender struct{ queue *message_queue.Queue }

type message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NewModule 创建邮件模块，默认同步发送；如需异步，由调用方注入队列发送器。
func NewModule(cfg config.MailConfig) *Module {
	return &Module{Sender: &provider{cfg: cfg}}
}

// Consume 返回队列消费 handler：解析消息后通过 sender 实际发送。
func Consume(sender Sender) message_queue.Handler {
	return func(ctx context.Context, msg message_queue.Message) error {
		data, err := json.Marshal(msg.Body)
		if err != nil {
			return fmt.Errorf("marshal mail message: %w", err)
		}
		var item message
		if err := json.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("unmarshal mail message: %w", err)
		}
		return sender.Send(ctx, item.To, item.Subject, item.Body)
	}
}

// NewQueueSender 返回通过队列异步发送的 Sender。
func NewQueueSender(queue *message_queue.Queue) Sender {
	return &queueSender{queue: queue}
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
