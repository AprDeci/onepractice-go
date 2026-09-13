package service_test

import (
	"encoding/json"
	"testing"

	"onepractice-golang/internal/agent"
	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/service"
)

func TestBuildEssayResultNilResult(t *testing.T) {
	task := &dto.EssayTask{
		ID:       "task-1",
		UserID:   42,
		RecordID: "record-1",
		Input:    agent.Input{Title: "标题"},
	}

	got := service.BuildEssayResult(task)

	if got.TaskID != "task-1" {
		t.Fatalf("TaskID = %q, want task-1", got.TaskID)
	}
	if got.UserID != 42 {
		t.Fatalf("UserID = %d, want 42", got.UserID)
	}
	if got.RecordID != "record-1" {
		t.Fatalf("RecordID = %q, want record-1", got.RecordID)
	}
	if got.Title != "标题" {
		t.Fatalf("Title = %q, want 标题", got.Title)
	}
	if got.FullScore != 0 || got.TotalScore != 0 || got.GrammarScore != 0 ||
		got.TopicScore != 0 || got.WordScore != 0 || got.StructureScore != 0 || got.WordNum != 0 {
		t.Fatalf("scores = %+v, want all zero", got)
	}
	if got.RawResult != "" {
		t.Fatalf("RawResult = %q, want empty", got.RawResult)
	}
}

func TestBuildEssayResultWithOutput(t *testing.T) {
	out := agent.Output{
		Title:      "标题",
		WordNum:    123,
		FullScore:  15,
		TotalScore: 12.5,
	}
	out.MajorScore.GrammarScore = 3.5
	out.MajorScore.TopicScore = 3
	out.MajorScore.WordScore = 3
	out.MajorScore.StructureScore = 3

	task := &dto.EssayTask{
		ID:       "task-2",
		UserID:   7,
		RecordID: "record-2",
		Input:    agent.Input{Title: "标题"},
		Result:   &out,
	}

	got := service.BuildEssayResult(task)

	if got.FullScore != 15 {
		t.Fatalf("FullScore = %d, want 15", got.FullScore)
	}
	if got.TotalScore != 12.5 {
		t.Fatalf("TotalScore = %v, want 12.5", got.TotalScore)
	}
	if got.GrammarScore != 3.5 {
		t.Fatalf("GrammarScore = %v, want 3.5", got.GrammarScore)
	}
	if got.TopicScore != 3 || got.WordScore != 3 || got.StructureScore != 3 {
		t.Fatalf("分项分错误: %+v", got)
	}
	if got.WordNum != 123 {
		t.Fatalf("WordNum = %d, want 123", got.WordNum)
	}
	if got.RawResult == "" {
		t.Fatal("RawResult 为空，希望非空合法 JSON")
	}

	var decoded agent.Output
	if err := json.Unmarshal([]byte(got.RawResult), &decoded); err != nil {
		t.Fatalf("RawResult 非法 JSON: %v", err)
	}
	if decoded.WordNum != 123 || decoded.TotalScore != 12.5 {
		t.Fatalf("RawResult 回解结果错误: %+v", decoded)
	}
}
