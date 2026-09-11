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
	RecordID  string        `json:"recordId,omitempty"`
	Status    string        `json:"status"`
	Input     agent.Input   `json:"input"`
	Result    *agent.Output `json:"result,omitempty"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// CreateEssayTaskRequest 是创建作文批改任务的请求体。
type CreateEssayTaskRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     string `json:"type" binding:"required"`
	RecordID string `json:"recordId"`
}

// EssayResultResponse 是作文评分结果的对外返回结构。
type EssayResultResponse struct {
	TaskID         string        `json:"taskId"`
	RecordID       string        `json:"recordId,omitempty"`
	Title          string        `json:"title"`
	FullScore      int           `json:"fullScore"`
	TotalScore     float64       `json:"totalScore"`
	GrammarScore   float64       `json:"grammarScore"`
	TopicScore     float64       `json:"topicScore"`
	WordScore      float64       `json:"wordScore"`
	StructureScore float64       `json:"structureScore"`
	WordNum        int           `json:"wordNum"`
	Result         *agent.Output `json:"result,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
}
