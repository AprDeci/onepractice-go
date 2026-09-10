package service_test

import (
	"context"
	"errors"
	"testing"

	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/common/message_queue"
	"onepractice-golang/internal/service"

	"github.com/redis/go-redis/v9"
)

// deadRedis 返回一个指向不可达地址的客户端，用于触发连接/上下文错误。
func deadRedis(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestEssayServiceWithoutRedisReturnsDisabled(t *testing.T) {
	svc := service.NewEssayService(nil, nil)

	if _, err := svc.CreateTask(context.Background(), 1, agent.Input{Title: "t", Content: "c", Type: "四级"}); !errors.Is(err, service.ErrRedisDisabled) {
		t.Fatalf("CreateTask() error = %v, want ErrRedisDisabled", err)
	}
	if _, err := svc.GetTask(context.Background(), 1, "task"); !errors.Is(err, service.ErrRedisDisabled) {
		t.Fatalf("GetTask() error = %v, want ErrRedisDisabled", err)
	}
}

func TestEssayServiceCreateTaskPropagatesCanceledContext(t *testing.T) {
	svc := service.NewEssayService(deadRedis(t), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := svc.CreateTask(ctx, 1, agent.Input{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateTask() error = %v, want context.Canceled", err)
	}
}

func TestEssayServiceHandleRejectsMalformedJob(t *testing.T) {
	svc := service.NewEssayService(deadRedis(t), nil)

	if err := svc.Handle(context.Background(), message_queue.Message{Body: 123}); err == nil {
		t.Fatal("Handle() error = nil, want unmarshal error")
	}
}

func TestEssayServiceHandleSkipsEmptyTaskID(t *testing.T) {
	svc := service.NewEssayService(deadRedis(t), nil)

	if err := svc.Handle(context.Background(), message_queue.Message{Body: map[string]string{"taskId": ""}}); err != nil {
		t.Fatalf("Handle() error = %v, want nil", err)
	}
}

func TestEssayServiceHandlePropagatesCanceledContext(t *testing.T) {
	svc := service.NewEssayService(deadRedis(t), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.Handle(ctx, message_queue.Message{Body: map[string]string{"taskId": "task-1"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handle() error = %v, want context.Canceled", err)
	}
}
