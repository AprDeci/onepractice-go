package service

import (
	"sort"
	"strconv"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/model"

	"gorm.io/gorm"
)

type QuestionService struct {
	db *gorm.DB
}

func NewQuestionService(db *gorm.DB) *QuestionService {
	return &QuestionService{db: db}
}

func (s *QuestionService) ByPaperID(paperID int) ([]model.Question, error) {
	if s.db == nil {
		return nil, ErrDatabaseDisabled
	}

	var questions []model.Question
	err := s.db.Where("paper_id = ?", paperID).Find(&questions).Error
	return questions, err
}

func (s *QuestionService) ByPaperIDAndType(paperID int, questionType string) ([]model.Question, error) {
	if s.db == nil {
		return nil, ErrDatabaseDisabled
	}

	var questions []model.Question
	err := s.db.Where("paper_id = ? and question_type = ?", paperID, questionType).Find(&questions).Error
	return questions, err
}

func (s *QuestionService) SplitByPart(paperID int) (dto.ExamQuestion, error) {
	if s.db == nil {
		return dto.ExamQuestion{}, ErrDatabaseDisabled
	}

	var questions []model.Question
	err := s.db.Where("paper_id = ?", paperID).
		Order("part_name asc, section_name asc, question_order asc").
		Find(&questions).Error
	if err != nil {
		return dto.ExamQuestion{}, err
	}

	parts := make([]dto.QuestionPart, 0)
	partIndex := make(map[string]int)
	for _, question := range questions {
		idx, ok := partIndex[question.PartName]
		if !ok {
			idx = len(parts)
			partIndex[question.PartName] = idx
			parts = append(parts, dto.QuestionPart{})
		}
		parts[idx].Questions = append(parts[idx].Questions, question)
	}

	return dto.ExamQuestion{PaperID: paperID, QuestionParts: parts}, nil
}

func (s *QuestionService) Answers(paperID int) (dto.AnswersResponse, error) {
	if s.db == nil {
		return dto.AnswersResponse{}, ErrDatabaseDisabled
	}

	var questions []model.Question
	err := s.db.Select("correct_answer").Where("paper_id = ?", paperID).Find(&questions).Error
	if err != nil {
		return dto.AnswersResponse{}, err
	}

	answers := make(model.Answers, 0)
	for _, question := range questions {
		answers = append(answers, question.CorrectAnswer...)
	}
	sort.Slice(answers, func(i, j int) bool {
		left, leftErr := strconv.Atoi(answers[i].Index)
		right, rightErr := strconv.Atoi(answers[j].Index)
		if leftErr == nil && rightErr == nil {
			return left < right
		}
		return answers[i].Index < answers[j].Index
	})

	return dto.AnswersResponse{PaperID: paperID, Answers: answers}, nil
}
