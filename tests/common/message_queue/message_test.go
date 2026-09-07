package message_queue_test

import (
	"testing"
	"time"

	"onepractice-golang/internal/common/message_queue"
)

func TestNewMessageDefaultsMaxRetries(t *testing.T) {
	msg := message_queue.NewMessage("", time.Now(), map[string]string{"type": "test"})
	if msg.MaxRetries != 5 {
		t.Fatalf("MaxRetries = %d, want 5", msg.MaxRetries)
	}
	if msg.RetryCount != 0 {
		t.Fatalf("RetryCount = %d, want 0", msg.RetryCount)
	}
	if msg.Id == "" {
		t.Fatal("Id is empty")
	}
}
