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

type PracticeQuestionRequest struct {
	QuestionType string `json:"questionType" binding:"required"`
	UnitCount    int    `json:"unitCount" binding:"required"`
}

type PracticeQuestionGroup struct {
	PaperID       int            `json:"paperId"`
	QuestionParts []QuestionPart `json:"questionParts"`
	Answers       model.Answers  `json:"answers"`
}

type PracticeQuestionResponse struct {
	QuestionType string                  `json:"questionType"`
	Groups       []PracticeQuestionGroup `json:"groups"`
}
