package service

import (
	"errors"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/model"

	"gorm.io/gorm"
)

type PaperService struct {
	db *gorm.DB
}

func NewPaperService(db *gorm.DB) *PaperService {
	return &PaperService{db: db}
}

func (s *PaperService) All() ([]model.Paper, error) {
	if s.db == nil {
		return nil, ErrDatabaseDisabled
	}

	var papers []model.Paper
	err := s.db.Order("exam_year desc, exam_month desc").Find(&papers).Error
	return papers, err
}

func (s *PaperService) Page(query dto.PaperQueryRequest) (dto.PageResult[model.Paper], error) {
	if s.db == nil {
		return dto.PageResult[model.Paper]{}, ErrDatabaseDisabled
	}

	db := s.paperFilter(s.db.Table("papers p"), query)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return dto.PageResult[model.Paper]{}, err
	}

	var papers []model.Paper
	err := db.Select(`p.paper_id, p.paper_name, p.exam_year, p.exam_month, p.version, p.total_time, p.type,
		(select count(*) from questions q where q.paper_id = p.paper_id) as question_count`).
		Order("p.exam_year desc, p.exam_month desc").
		Limit(query.Size).
		Offset(offset(query.Page, query.Size)).
		Scan(&papers).Error

	return dto.PageResult[model.Paper]{Total: total, Data: papers}, err
}

func (s *PaperService) PageWithRating(query dto.PaperQueryRequest) (dto.PageResult[dto.PaperWithRating], error) {
	if s.db == nil {
		return dto.PageResult[dto.PaperWithRating]{}, ErrDatabaseDisabled
	}

	base := s.paperFilter(s.db.Table("papers p"), query).
		Joins("left join paper_rate_mapping r on r.paperId = p.paper_id").
		Joins("left join questions q on q.paper_id = p.paper_id").
		Group("p.paper_id, p.paper_name, p.exam_year, p.exam_month, p.version, p.total_time, p.type, r.rating, r.number").
		Having("count(q.question_id) > 0")

	var totalRows []struct{ PaperID int }
	if err := base.Select("p.paper_id").Scan(&totalRows).Error; err != nil {
		return dto.PageResult[dto.PaperWithRating]{}, err
	}

	var papers []dto.PaperWithRating
	err := base.Select(`p.paper_id, p.paper_name, p.exam_year, p.exam_month, p.version, p.total_time, p.type,
		count(q.question_id) as question_count, coalesce(r.rating, 0) as rating, coalesce(r.number, 0) as number`).
		Order("p.exam_year desc, p.exam_month desc, substring(p.type, 4, 1) desc").
		Limit(query.Size).
		Offset(offset(query.Page, query.Size)).
		Scan(&papers).Error

	return dto.PageResult[dto.PaperWithRating]{Total: int64(len(totalRows)), Data: papers}, err
}

func (s *PaperService) ByType(paperType string) ([]model.Paper, error) {
	if s.db == nil {
		return nil, ErrDatabaseDisabled
	}

	var papers []model.Paper
	err := s.db.Where("type = ?", paperType).Find(&papers).Error
	return papers, err
}

func (s *PaperService) Types() ([]string, error) {
	if s.db == nil {
		return nil, ErrDatabaseDisabled
	}

	var types []string
	err := s.db.Model(&model.Paper{}).Distinct().Pluck("type", &types).Error
	return types, err
}

func (s *PaperService) Intro(id int) (dto.PaperIntro, error) {
	if s.db == nil {
		return dto.PaperIntro{}, ErrDatabaseDisabled
	}

	var paper model.Paper
	err := s.db.Where("paper_id = ?", id).First(&paper).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PaperIntro{}, err
		}
		return dto.PaperIntro{}, err
	}

	return dto.PaperIntro{
		PaperName:            paper.PaperName,
		ExamYear:             paper.ExamYear,
		ExamMonth:            paper.ExamMonth,
		PaperType:            paper.Type,
		PaperTime:            paper.TotalTime,
		Difficulty:           "",
		SectionCount:         partCount(paper.Type),
		SectionQuestionCount: partQuestionCount(paper.Type),
	}, nil
}

func (s *PaperService) paperFilter(db *gorm.DB, query dto.PaperQueryRequest) *gorm.DB {
	if query.Type != "" {
		db = db.Where("p.type = ?", query.Type)
	}
	if query.Year != 0 {
		db = db.Where("p.exam_year = ?", query.Year)
	}
	return db
}

func offset(page, size int) int {
	return (page - 1) * size
}

func partCount(paperType string) int64 {
	switch paperType {
	case "CET-4", "CET-6":
		return 4
	default:
		return 0
	}
}

func partQuestionCount(paperType string) []int {
	switch paperType {
	case "CET-4", "CET-6":
		return []int{1, 20, 30, 1}
	default:
		return nil
	}
}
