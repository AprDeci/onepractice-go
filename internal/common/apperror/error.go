package apperror

import "fmt"

// Error 描述可安全返回给客户端的业务错误，并保留底层原因供内部判断。
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// New 创建不包含底层原因的业务错误。
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap 使用业务错误包装底层错误，保留 errors.Is 和 errors.As 语义。
func Wrap(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

// Error 返回包含底层原因的内部错误文本。
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// Unwrap 返回被包装的底层错误。
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
