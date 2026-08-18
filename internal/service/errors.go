package service

import "errors"

var ErrDatabaseDisabled = errors.New("database disabled: MYSQL_DSN is empty")

var ErrRedisDisabled = errors.New("redis disabled: REDIS_DISABLED=true or redis unavailable")

var ErrInvalidQuestionType = errors.New("invalid question type")

var ErrInvalidPracticeUnitCount = errors.New("invalid practice unit count")

var ErrPracticeQuestionsNotFound = errors.New("no practice questions found")
