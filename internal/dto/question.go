package dto

type PracticeQuestionRequest struct {
	QuestionType string `json:"questionType" binding:"required"`
	UnitCount    int    `json:"unitCount" binding:"required"`
}
