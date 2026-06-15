package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"onepractice-golang/internal/dto"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	recordTTL       = 30 * 24 * time.Hour
	defaultListDays = 30
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

type RecordService struct {
	redis *redis.Client
	paper *PaperService
}

func NewRecordService(redisClient *redis.Client, paperService *PaperService) *RecordService {
	return &RecordService{redis: redisClient, paper: paperService}
}

func (s *RecordService) Create(userID int64, req dto.RecordRequest) (string, error) {
	if s.redis == nil {
		return "", ErrRedisDisabled
	}
	if req.PaperID == 0 {
		return "", ErrInvalidParam
	}

	paperID := int(req.PaperID)
	intro, err := s.paper.Intro(paperID)
	if err != nil {
		return "", ErrInvalidParam
	}

	now := time.Now().UnixMilli()
	recordID := strings.ReplaceAll(uuid.NewString(), "-", "")
	record := dto.UserExamRecord{
		RecordID:     recordID,
		UserID:       userID,
		PaperID:      paperID,
		PaperType:    intro.PaperType,
		PaperName:    formatPaperName(intro),
		Type:         req.Type,
		IsFinished:   req.IsFinished,
		Answers:      req.Answers,
		TimeSpend:    req.TimeSpend,
		Score:        req.Score,
		TotalScore:   req.TotalScore,
		Timestamp:    now,
		HasSpendTime: int64(req.HasSpendTime),
	}
	if err := s.saveRecord(context.Background(), record); err != nil {
		return "", err
	}
	return recordID, nil
}

func (s *RecordService) ListRecent(userID int64, days, pageNum, pageSize int) ([]dto.UserExamRecord, error) {
	if s.redis == nil {
		return nil, ErrRedisDisabled
	}
	if days <= 0 {
		days = defaultListDays
	}
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	ctx := context.Background()
	minTimestamp := time.Now().Add(-time.Duration(days) * 24 * time.Hour).UnixMilli()
	zsetKey := recordSortedSetKey(userID)
	_ = s.redis.ZRemRangeByScore(ctx, zsetKey, "-inf", strconv.FormatInt(minTimestamp-1, 10)).Err()
	_ = s.redis.Expire(ctx, zsetKey, recordTTL).Err()

	start := int64((pageNum - 1) * pageSize)
	stop := start + int64(pageSize) - 1
	recordIDs, err := s.redis.ZRevRange(ctx, zsetKey, start, stop).Result()
	if err != nil {
		return nil, err
	}

	records := make([]dto.UserExamRecord, 0, len(recordIDs))
	missing := make([]string, 0)
	for _, recordID := range recordIDs {
		payload, getErr := s.redis.Get(ctx, recordDetailKey(userID, recordID)).Result()
		if getErr != nil {
			missing = append(missing, recordID)
			continue
		}
		var record dto.UserExamRecord
		if unmarshalErr := json.Unmarshal([]byte(payload), &record); unmarshalErr != nil {
			missing = append(missing, recordID)
			continue
		}
		records = append(records, record)
	}
	if len(missing) > 0 {
		members := make([]redis.Z, 0, len(missing))
		_ = members
		args := make([]interface{}, 0, len(missing))
		for _, recordID := range missing {
			args = append(args, recordID)
		}
		_ = s.redis.ZRem(ctx, zsetKey, args...).Err()
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Timestamp > records[j].Timestamp })
	return records, nil
}

func (s *RecordService) Update(userID int64, req dto.RecordRequest) error {
	if s.redis == nil {
		return ErrRedisDisabled
	}
	if req.RecordID == "" {
		return ErrInvalidParam
	}

	ctx := context.Background()
	payload, err := s.redis.Get(ctx, recordDetailKey(userID, req.RecordID)).Result()
	if err != nil {
		return ErrInvalidParam
	}

	var record dto.UserExamRecord
	if err = json.Unmarshal([]byte(payload), &record); err != nil {
		return ErrInvalidParam
	}
	record.Timestamp = time.Now().UnixMilli()
	record.Score = req.Score
	record.Answers = req.Answers
	record.IsFinished = req.IsFinished
	record.HasSpendTime = int64(req.HasSpendTime)
	return s.saveRecord(ctx, record)
}

func (s *RecordService) saveRecord(ctx context.Context, record dto.UserExamRecord) error {
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	pipe := s.redis.Pipeline()
	detailKey := recordDetailKey(record.UserID, record.RecordID)
	zsetKey := recordSortedSetKey(record.UserID)
	pipe.Set(ctx, detailKey, payload, recordTTL)
	pipe.ZAdd(ctx, zsetKey, redis.Z{Score: float64(record.Timestamp), Member: record.RecordID})
	pipe.Expire(ctx, zsetKey, recordTTL)
	pipe.ZRemRangeByScore(ctx, zsetKey, "-inf", strconv.FormatInt(time.Now().Add(-recordTTL).UnixMilli(), 10))
	_, err = pipe.Exec(ctx)
	return err
}

func recordDetailKey(userID int64, recordID string) string {
	return fmt.Sprintf("onepractice:record:user:%d:%s", userID, recordID)
}

func recordSortedSetKey(userID int64) string {
	return fmt.Sprintf("onepractice:user:record:sorted:%d", userID)
}

func formatPaperName(intro dto.PaperIntro) string {
	return fmt.Sprintf("%d年%d月%s", intro.ExamYear, intro.ExamMonth, intro.PaperName)
}
