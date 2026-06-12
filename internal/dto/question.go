package dto

import "onepractice-golang/internal/model"

type QuestionPart struct {
	Questions []model.Question `json:"questions"`
}

type ExamQuestion struct {
	PaperID       int            `json:"paperId"`
	QuestionParts []QuestionPart `json:"questionParts"`
}

type AnswersResponse struct {
	PaperID int           `json:"paperId"`
	Answers model.Answers `json:"answers"`
}
