package service

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
)

// 依赖不可用时只向客户端暴露统一的 503 文案，内部诊断串保留在包装原因中（仅进日志）。
var ErrDatabaseDisabled = apperror.Wrap(apperror.CodeServiceUnavailable, "依赖服务不可用", errors.New("database disabled: MYSQL_DSN is empty"))

var ErrRedisDisabled = apperror.Wrap(apperror.CodeServiceUnavailable, "依赖服务不可用", errors.New("redis disabled: REDIS_DISABLED=true or redis unavailable"))

var ErrInvalidQuestionType = errors.New("invalid question type")

var ErrInvalidPracticeUnitCount = errors.New("invalid practice unit count")

var ErrPracticeQuestionsNotFound = errors.New("no practice questions found")

var ErrTaskNotFound = errors.New("task not found")
