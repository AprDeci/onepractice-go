package service_test

import (
	"errors"
	"testing"

	"onepractice-golang/internal/service"
)

func TestEssayServiceGetResultsByRecordWithoutDatabaseReturnsDisabled(t *testing.T) {
	svc := service.NewEssayService(nil, nil, nil)

	if _, err := svc.GetResultsByRecord(1, "record-1"); !errors.Is(err, service.ErrDatabaseDisabled) {
		t.Fatalf("GetResultsByRecord() error = %v, want ErrDatabaseDisabled", err)
	}
}
