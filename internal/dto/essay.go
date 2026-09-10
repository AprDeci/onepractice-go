package dto

import (
	"time"

	"onepractice-golang/internal/agent"
)

// 作文批改任务状态。
const (
	EssayTaskPending    = "pending"
	EssayTaskProcessing = "processing"
	EssayTaskSucceeded  = "succeeded"
	EssayTaskFailed     = "failed"
)

// EssayTask 是作文批改异步任务的状态与结果。
type EssayTask struct {
	ID        string        `json:"taskId"`
	UserID    int64         `json:"userId"`
	Status    string        `json:"status"`
	Input     agent.Input   `json:"input"`
	Result    *agent.Output `json:"result,omitempty"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// CreateEssayTaskRequest 是创建作文批改任务的请求体。
type CreateEssayTaskRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"`
}
