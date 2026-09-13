package dto

// 作文批改任务状态。
const (
	EssayTaskPending    = "pending"
	EssayTaskProcessing = "processing"
	EssayTaskSucceeded  = "succeeded"
	EssayTaskFailed     = "failed"
)

// CreateEssayTaskRequest 是创建作文批改任务的请求体。
type CreateEssayTaskRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     string `json:"type" binding:"required"`
	RecordID string `json:"recordId"`
}
