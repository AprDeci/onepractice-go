package service

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
)

var ErrDatabaseDisabled = apperror.Wrap(apperror.CodeServiceUnavailable, "依赖服务不可用", errors.New("database disabled: MYSQL_DSN is empty"))

var ErrRedisDisabled = apperror.Wrap(apperror.CodeServiceUnavailable, "依赖服务不可用", errors.New("redis disabled: REDIS_DISABLED=true or redis unavailable"))

var ErrInvalidQuestionType = apperror.New(apperror.CodeInvalidArgument, "参数无效")

var ErrInvalidPracticeUnitCount = apperror.New(apperror.CodeInvalidArgument, "参数无效")

var ErrPracticeQuestionsNotFound = apperror.New(apperror.CodeNotFound, "资源不存在")

var ErrTaskNotFound = apperror.New(apperror.CodeNotFound, "任务不存在")
