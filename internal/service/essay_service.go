package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/common/message_queue"
	"onepractice-golang/internal/dto"
	appmodel "onepractice-golang/internal/model"

	"github.com/cloudwego/eino/components/model"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	essayTaskTTL = 24 * time.Hour
	// EssayTopic 是作文批改任务队列的 topic。
	EssayTopic          = "onepractice:essay:grade"
	essayProcessTimeout = 120 * time.Second
	essayStaleAfter     = 10 * time.Minute
	essayTaskKeyPrefix  = "onepractice:essay:task:"
)

// EssayService 负责作文批改任务的创建、异步执行与查询。
type EssayService struct {
	redis *redis.Client
	model model.BaseChatModel
	queue *message_queue.Queue
	db    *gorm.DB
}

// NewEssayService 创建作文批改服务。redis 或 model 为 nil 时仅创建空壳，
// 相关方法会返回 ErrRedisDisabled。
func NewEssayService(redisClient *redis.Client, cm model.BaseChatModel, db *gorm.DB) *EssayService {
	return &EssayService{redis: redisClient, model: cm, db: db}
}

// SetQueue 注入用于投递批改任务的队列。
func (s *EssayService) SetQueue(queue *message_queue.Queue) {
	s.queue = queue
}

type essayJob struct {
	TaskID   string `json:"taskId"`
	UserID   int64  `json:"userId"`
	RecordID string `json:"recordId,omitempty"`
}

// CreateTask 落库任务并投递到队列，返回任务 ID。
func (s *EssayService) CreateTask(ctx context.Context, userID int64, recordID string, input agent.Input) (string, error) {
	if s == nil || s.redis == nil {
		return "", ErrRedisDisabled
	}

	taskID := strings.ReplaceAll(uuid.NewString(), "-", "")
	now := time.Now()
	task := dto.EssayTask{
		ID:        taskID,
		UserID:    userID,
		RecordID:  recordID,
		Status:    dto.EssayTaskPending,
		Input:     input,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.saveTask(ctx, &task); err != nil {
		return "", err
	}
	if s.queue != nil {
		if _, err := s.queue.Publish(message_queue.NewMessage("", now, essayJob{TaskID: taskID, UserID: userID, RecordID: recordID})); err != nil {
			return "", err
		}
	}
	return taskID, nil
}

// GetTask 查询任务，校验归属，并对长时间停留在 processing 的任务做超时兜底。
func (s *EssayService) GetTask(ctx context.Context, userID int64, taskID string) (*dto.EssayTask, error) {
	if s == nil || s.redis == nil {
		return nil, ErrRedisDisabled
	}

	task, err := s.loadTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil || task.UserID != userID {
		return nil, ErrTaskNotFound
	}

	if task.Status == dto.EssayTaskProcessing && time.Since(task.UpdatedAt) > essayStaleAfter {
		task.Status = dto.EssayTaskFailed
		task.Error = "处理超时"
		task.UpdatedAt = time.Now()
		if err := s.saveTask(ctx, task); err != nil {
			return nil, err
		}
	}
	return task, nil
}

// Handle 消费批改任务消息：加载任务、调用模型批改并回写结果。
func (s *EssayService) Handle(ctx context.Context, msg message_queue.Message) error {
	data, err := json.Marshal(msg.Body)
	if err != nil {
		return fmt.Errorf("marshal essay job: %w", err)
	}
	var job essayJob
	if err := json.Unmarshal(data, &job); err != nil {
		return fmt.Errorf("unmarshal essay job: %w", err)
	}
	if job.TaskID == "" {
		return nil
	}

	task, err := s.loadTask(ctx, job.TaskID)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	if task.Status == dto.EssayTaskSucceeded || task.Status == dto.EssayTaskFailed {
		return nil
	}

	task.Status = dto.EssayTaskProcessing
	task.UpdatedAt = time.Now()
	if err := s.saveTask(ctx, task); err != nil {
		return err
	}

	procCtx, cancel := context.WithTimeout(ctx, essayProcessTimeout)
	defer cancel()
	out, scoreErr := agent.EssayScore(procCtx, s.model, task.Input)

	task.UpdatedAt = time.Now()
	if scoreErr != nil {
		task.Status = dto.EssayTaskFailed
		task.Error = scoreErr.Error()
		_ = s.saveTask(ctx, task)
		return nil
	}

	task.Status = dto.EssayTaskSucceeded
	task.Result = &out
	if err := s.saveTask(ctx, task); err != nil {
		return err
	}
	if s.db != nil {
		result := BuildEssayResult(task)
		if err := result.Upsert(s.db); err != nil {
			slog.Warn("持久化作文评分结果失败", slog.String("taskId", task.ID), slog.Any("err", err))
		}
	}
	return nil
}

// BuildEssayResult 将作文任务及其评分输出映射为持久化结构。
func BuildEssayResult(task *dto.EssayTask) appmodel.EssayGradingResult {
	result := appmodel.EssayGradingResult{
		TaskID:   task.ID,
		UserID:   task.UserID,
		RecordID: task.RecordID,
		Title:    task.Input.Title,
	}
	if task.Result == nil {
		return result
	}
	out := task.Result
	result.FullScore = out.FullScore
	result.TotalScore = out.TotalScore
	result.GrammarScore = out.MajorScore.GrammarScore
	result.TopicScore = out.MajorScore.TopicScore
	result.WordScore = out.MajorScore.WordScore
	result.StructureScore = out.MajorScore.StructureScore
	result.WordNum = out.WordNum
	if raw, err := json.Marshal(out); err == nil {
		result.RawResult = string(raw)
	}
	return result
}

// GetResultsByRecord 查询某次考试记录关联的作文评分结果，按评分时间倒序。
func (s *EssayService) GetResultsByRecord(userID int64, recordID string) ([]dto.EssayResultResponse, error) {
	if s == nil || s.db == nil {
		return nil, ErrDatabaseDisabled
	}
	rows, err := appmodel.ListEssayResultsByRecord(s.db, userID, recordID)
	if err != nil {
		return nil, err
	}
	results := make([]dto.EssayResultResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, toEssayResultResponse(row))
	}
	return results, nil
}

// ListStandaloneResults 分页查询当前用户的独立作文评分历史（不关联考试记录）。
func (s *EssayService) ListStandaloneResults(userID int64, page, pageSize int) ([]dto.EssayResultResponse, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, ErrDatabaseDisabled
	}
	offset := (page - 1) * pageSize
	rows, total, err := appmodel.ListStandaloneEssayResults(s.db, userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	results := make([]dto.EssayResultResponse, 0, len(rows))
	for _, row := range rows {
		results = append(results, toEssayResultResponse(row))
	}
	return results, total, nil
}

// GetResultByTask 从数据库查询某任务的作文评分结果，不受 Redis TTL 影响。
func (s *EssayService) GetResultByTask(userID int64, taskID string) (*dto.EssayResultResponse, error) {
	if s == nil || s.db == nil {
		return nil, ErrDatabaseDisabled
	}
	row, err := appmodel.GetEssayResultByTask(s.db, userID, taskID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrTaskNotFound
	}
	resp := toEssayResultResponse(*row)
	return &resp, nil
}

func toEssayResultResponse(row appmodel.EssayGradingResult) dto.EssayResultResponse {
	resp := dto.EssayResultResponse{
		TaskID:         row.TaskID,
		RecordID:       row.RecordID,
		Title:          row.Title,
		FullScore:      row.FullScore,
		TotalScore:     row.TotalScore,
		GrammarScore:   row.GrammarScore,
		TopicScore:     row.TopicScore,
		WordScore:      row.WordScore,
		StructureScore: row.StructureScore,
		WordNum:        row.WordNum,
		CreatedAt:      row.CreatedAt,
	}
	if row.RawResult != "" {
		var out agent.Output
		if err := json.Unmarshal([]byte(row.RawResult), &out); err == nil {
			resp.Result = &out
		}
	}
	return resp
}

// 保存任务
func (s *EssayService) saveTask(ctx context.Context, task *dto.EssayTask) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, essayTaskKeyPrefix+task.ID, payload, essayTaskTTL).Err()
}

func (s *EssayService) loadTask(ctx context.Context, taskID string) (*dto.EssayTask, error) {
	payload, err := s.redis.Get(ctx, essayTaskKeyPrefix+taskID).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var task dto.EssayTask
	if err := json.Unmarshal(payload, &task); err != nil {
		return nil, err
	}
	return &task, nil
}
