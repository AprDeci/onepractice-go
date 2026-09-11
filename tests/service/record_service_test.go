package service_test

import (
	"errors"
	"testing"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/service"
)

func TestRecordServiceWithoutDatabaseReturnsDisabled(t *testing.T) {
	svc := service.NewRecordService(nil, nil)

	if _, err := svc.Create(1, dto.RecordRequest{PaperID: 1}); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Create() error = %v, want ErrDatabaseDisabled", err)
	}
	if _, _, err := svc.ListRecentPage(1, 30, 1, 10); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("ListRecentPage() error = %v, want ErrDatabaseDisabled", err)
	}
	if err := svc.Update(1, dto.RecordRequest{RecordID: "x"}); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("Update() error = %v, want ErrDatabaseDisabled", err)
	}
}
