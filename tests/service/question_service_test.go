package service_test

import (
	"errors"
	"testing"

	"onepractice-golang/internal/service"
)

func TestQuestionServicePracticeNilDB(t *testing.T) {
	questionService := service.NewQuestionService(nil)

	for _, questionType := range []string{"listening", "cloze", "matching", "reading"} {
		t.Run(questionType, func(t *testing.T) {
			_, err := questionService.Practice(questionType, 1)
			if !errors.Is(err, service.ErrDatabaseDisabled) {
				t.Fatalf("Practice(%q, 1) error = %v, want %v", questionType, err, service.ErrDatabaseDisabled)
			}
		})
	}
}

func TestQuestionServicePracticeInvalidTypeBeforeNilDB(t *testing.T) {
	questionService := service.NewQuestionService(nil)

	for _, questionType := range []string{"", "essay"} {
		t.Run(questionType, func(t *testing.T) {
			_, err := questionService.Practice(questionType, 1)
			if !errors.Is(err, service.ErrInvalidQuestionType) {
				t.Fatalf("Practice(%q, 1) error = %v, want %v", questionType, err, service.ErrInvalidQuestionType)
			}
		})
	}
}

func TestQuestionServicePracticeInvalidUnitCountBeforeNilDB(t *testing.T) {
	questionService := service.NewQuestionService(nil)

	for _, unitCount := range []int{0, 2, 4, 6} {
		t.Run("invalid", func(t *testing.T) {
			_, err := questionService.Practice("listening", unitCount)
			if !errors.Is(err, service.ErrInvalidPracticeUnitCount) {
				t.Fatalf("Practice(listening, %d) error = %v, want %v", unitCount, err, service.ErrInvalidPracticeUnitCount)
			}
		})
	}
}
