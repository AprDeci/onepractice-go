package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EssayGradingResult 是作文 AI 评分结果的持久化形态；record_id 为空表示独立作文批改，
// 非空表示关联到某次考试记录。
type EssayGradingResult struct {
	gorm.Model
	TaskID         string  `gorm:"column:task_id" json:"taskId"`
	UserID         int64   `gorm:"column:user_id" json:"userId"`
	RecordID       string  `gorm:"column:record_id" json:"recordId"`
	Title          string  `gorm:"column:title" json:"title"`
	FullScore      int     `gorm:"column:full_score" json:"fullScore"`
	TotalScore     float64 `gorm:"column:total_score" json:"totalScore"`
	GrammarScore   float64 `gorm:"column:grammar_score" json:"grammarScore"`
	TopicScore     float64 `gorm:"column:topic_score" json:"topicScore"`
	WordScore      float64 `gorm:"column:word_score" json:"wordScore"`
	StructureScore float64 `gorm:"column:structure_score" json:"structureScore"`
	WordNum        int     `gorm:"column:word_num" json:"wordNum"`
	RawResult      string  `gorm:"column:raw_result;type:json" json:"rawResult"`
}

func (EssayGradingResult) TableName() string { return "essay_grading_results" }

// Upsert 以 task_id 为唯一键原子写入或更新，保证重复消费不产生重复结果。
func (g *EssayGradingResult) Upsert(db *gorm.DB) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "task_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"user_id", "record_id", "title", "full_score", "total_score",
			"grammar_score", "topic_score", "word_score", "structure_score",
			"word_num", "raw_result", "updated_at",
		}),
	}).Create(g).Error
}

// ListEssayResultsByRecord 查询某次考试记录关联的作文评分结果，按评分时间倒序。
func ListEssayResultsByRecord(db *gorm.DB, userID int64, recordID string) ([]EssayGradingResult, error) {
	var rows []EssayGradingResult
	err := db.Where("user_id = ? and record_id = ?", userID, recordID).
		Order("created_at desc").
		Find(&rows).Error
	return rows, err
}
