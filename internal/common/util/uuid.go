package util

import "github.com/google/uuid"

// NewRequestID 生成用于串联请求日志和响应的 UUID。
func NewRequestID() string {
	return uuid.NewString()
}
