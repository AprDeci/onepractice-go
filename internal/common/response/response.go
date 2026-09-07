package response

import (
	"errors"
	"net/http"

	"onepractice-golang/internal/common/apperror"

	"github.com/gin-gonic/gin"
)

// RequestIDKey 是请求 ID 中间件和响应工具共享的 Gin 上下文键。
const RequestIDKey = "request_id"

// Body 是所有 HTTP 处理器返回的稳定响应外壳。
type Body struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Success 写入成功响应，并带上当前请求 ID。
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{
		Code:      apperror.CodeOK,
		Message:   "ok",
		Data:      data,
		RequestID: requestID(c),
	})
}

// Error 将应用错误转换为统一响应结构。
func Error(c *gin.Context, err error) {
	if err != nil {
		c.Error(err)
	}
	status, body := errorBody(c, err)
	c.JSON(status, body)
}

func errorBody(c *gin.Context, err error) (int, Body) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return httpStatus(appErr.Code), Body{
			Code:      appErr.Code,
			Message:   appErr.Message,
			RequestID: requestID(c),
		}
	}

	return http.StatusInternalServerError, Body{
		Code:      apperror.CodeInternal,
		Message:   "系统异常",
		RequestID: requestID(c),
	}
}

// httpStatus 集中维护业务错误码到 HTTP 状态码的映射。
func httpStatus(code int) int {
	switch code {
	case apperror.CodeInvalidArgument:
		return http.StatusBadRequest
	case apperror.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func requestID(c *gin.Context) string {
	value, exists := c.Get(RequestIDKey)
	if !exists {
		return ""
	}
	id, _ := value.(string)
	return id
}
