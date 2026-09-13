package service

import (
	"fmt"
	"strings"
	"time"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultListDays = 30
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

type RecordService struct {
	paper *PaperService
	db    *gorm.DB
}

func NewRecordService(paperService *PaperService, db *gorm.DB) *RecordService {
	return &RecordService{paper: paperService, db: db}
}

func (s *RecordService) Create(userID int64, req dto.RecordRequest) (string, error) {
	if s.db == nil {
		return "", ErrDatabaseDisabled
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
	row := model.ExamRecord{
		RecordID:     recordID,
		UserID:       userID,
		PaperID:      paperID,
		PaperType:    intro.PaperType,
		PaperName:    formatPaperName(intro),
		ExamType:     req.Type,
		IsFinished:   req.IsFinished,
		Answers:      req.Answers,
		Score:        req.Score,
		TotalScore:   req.TotalScore,
		TimeSpend:    req.TimeSpend,
		HasSpendTime: int64(req.HasSpendTime),
		SubmitTS:     now,
	}
	if err := row.Upsert(s.db); err != nil {
		return "", err
	}
	return recordID, nil
}

func (s *RecordService) ListRecent(userID int64, days, pageNum, pageSize int) ([]dto.UserExamRecord, error) {
	records, _, err := s.ListRecentPage(userID, days, pageNum, pageSize)
	return records, err
}

func (s *RecordService) ListRecentPage(userID int64, days, pageNum, pageSize int) ([]dto.UserExamRecord, int64, error) {
	if s.db == nil {
		return nil, 0, ErrDatabaseDisabled
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

	sinceMS := time.Now().Add(-time.Duration(days) * 24 * time.Hour).UnixMilli()
	offset := (pageNum - 1) * pageSize
	rows, total, err := model.ListExamRecordsByUser(s.db, userID, sinceMS, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	records := make([]dto.UserExamRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, toUserExamRecord(row))
	}
	return records, total, nil
}

func (s *RecordService) Update(userID int64, req dto.RecordRequest) error {
	if s.db == nil {
		return ErrDatabaseDisabled
	}
	if req.RecordID == "" {
		return ErrInvalidParam
	}

	row, err := model.GetExamRecordByID(s.db, userID, req.RecordID)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrInvalidParam
	}
	row.Score = req.Score
	row.Answers = req.Answers
	row.IsFinished = req.IsFinished
	row.HasSpendTime = int64(req.HasSpendTime)
	row.SubmitTS = time.Now().UnixMilli()
	return row.Upsert(s.db)
}

func toUserExamRecord(row model.ExamRecord) dto.UserExamRecord {
	return dto.UserExamRecord{
		RecordID:     row.RecordID,
		UserID:       row.UserID,
		PaperID:      row.PaperID,
		PaperType:    row.PaperType,
		PaperName:    row.PaperName,
		Type:         row.ExamType,
		IsFinished:   row.IsFinished,
		Answers:      row.Answers,
		TimeSpend:    row.TimeSpend,
		Score:        row.Score,
		TotalScore:   row.TotalScore,
		Timestamp:    row.SubmitTS,
		HasSpendTime: row.HasSpendTime,
	}
}

func formatPaperName(intro dto.PaperIntro) string {
	return fmt.Sprintf("%d年%d月%s", intro.ExamYear, intro.ExamMonth, intro.PaperName)
}
