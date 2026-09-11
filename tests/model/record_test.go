package model_test

import (
	"testing"

	"onepractice-golang/internal/model"
)

func TestExamRecordTableName(t *testing.T) {
	if got := (model.ExamRecord{}).TableName(); got != "exam_records" {
		t.Fatalf("ExamRecord.TableName() = %q, want exam_records", got)
	}
}

func TestEssayGradingResultTableName(t *testing.T) {
	if got := (model.EssayGradingResult{}).TableName(); got != "essay_grading_results" {
		t.Fatalf("EssayGradingResult.TableName() = %q, want essay_grading_results", got)
	}
}
