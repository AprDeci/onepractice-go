package model

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ExamRecord 是考试作答记录在 MySQL 中的持久化形态。
type ExamRecord struct {
	gorm.Model
	RecordID     string `gorm:"column:record_id" json:"recordId"`
	UserID       int64  `gorm:"column:user_id" json:"userId"`
	PaperID      int    `gorm:"column:paper_id" json:"paperId"`
	PaperType    string `gorm:"column:paper_type" json:"paperType"`
	PaperName    string `gorm:"column:paper_name" json:"paperName"`
	ExamType     string `gorm:"column:exam_type" json:"examType"`
	IsFinished   int    `gorm:"column:is_finished" json:"isFinished"`
	Answers      string `gorm:"column:answers" json:"answers"`
	Score        int    `gorm:"column:score" json:"score"`
	TotalScore   int    `gorm:"column:total_score" json:"totalScore"`
	TimeSpend    int    `gorm:"column:time_spend" json:"timeSpend"`
	HasSpendTime int64  `gorm:"column:has_spend_time" json:"hasSpendTime"`
	SubmitTS     int64  `gorm:"column:submit_ts" json:"submitTs"`
}

func (ExamRecord) TableName() string { return "exam_records" }

// Upsert 以 record_id 为唯一键原子写入或更新，保证幂等。
func (r *ExamRecord) Upsert(db *gorm.DB) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "record_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"user_id", "paper_id", "paper_type", "paper_name", "exam_type",
			"is_finished", "answers", "score", "total_score", "time_spend",
			"has_spend_time", "submit_ts", "updated_at",
		}),
	}).Create(r).Error
}

// GetExamRecordByID 按用户与 record_id 查询单条记录；不存在时返回 (nil, nil)。
func GetExamRecordByID(db *gorm.DB, userID int64, recordID string) (*ExamRecord, error) {
	var row ExamRecord
	err := db.Where("user_id = ? and record_id = ?", userID, recordID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListExamRecordsByUser 按时间倒序分页查询某用户在指定时间之后的记录，同时返回总数。
func ListExamRecordsByUser(db *gorm.DB, userID int64, sinceMS int64, offset, limit int) ([]ExamRecord, int64, error) {
	query := db.Model(&ExamRecord{}).Where("user_id = ? and submit_ts >= ?", userID, sinceMS)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ExamRecord
	if err := query.Order("submit_ts desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
