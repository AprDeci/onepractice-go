package mail_test

import (
	"context"
	"testing"

	"onepractice-golang/internal/common/mail"
	"onepractice-golang/internal/common/message_queue"
	"onepractice-golang/internal/config"
)

type recordingSender struct {
	to      string
	subject string
	body    string
}

func (s *recordingSender) Send(_ context.Context, to, subject, body string) error {
	s.to, s.subject, s.body = to, subject, body
	return nil
}

func TestNewModuleDisabledSkipsSend(t *testing.T) {
	module := mail.NewModule(config.MailConfig{Disabled: true})

	if err := module.Sender.Send(context.Background(), "a@b.c", "s", "b"); err != nil {
		t.Fatalf("Send() error = %v, want nil when disabled", err)
	}
}

func TestNewModuleIncompleteConfigReturnsError(t *testing.T) {
	module := mail.NewModule(config.MailConfig{})

	if err := module.Sender.Send(context.Background(), "a@b.c", "s", "b"); err == nil {
		t.Fatal("Send() error = nil, want incomplete config error")
	}
}

func TestConsumeDispatchesToSender(t *testing.T) {
	sender := &recordingSender{}
	handler := mail.Consume(sender)

	err := handler(context.Background(), message_queue.Message{Body: map[string]string{
		"to":      "a@b.c",
		"subject": "hello",
		"body":    "world",
	}})
	if err != nil {
		t.Fatalf("Consume handler error = %v, want nil", err)
	}
	if sender.to != "a@b.c" || sender.subject != "hello" || sender.body != "world" {
		t.Fatalf("sender got to=%q subject=%q body=%q", sender.to, sender.subject, sender.body)
	}
}

func TestConsumePropagatesBodyMarshalError(t *testing.T) {
	handler := mail.Consume(&recordingSender{})

	if err := handler(context.Background(), message_queue.Message{Body: func() {}}); err == nil {
		t.Fatal("Consume handler error = nil, want marshal error")
	}
}

func TestQueueSenderWithoutQueueReturnsError(t *testing.T) {
	sender := mail.NewQueueSender(nil)

	if err := sender.Send(context.Background(), "a@b.c", "s", "b"); err == nil {
		t.Fatal("queue sender Send() error = nil, want nil queue error")
	}
}
