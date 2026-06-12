package response

type ErrorCode struct {
	Code int
	Msg  string
}

var (
	ErrPasswordError       = ErrorCode{10001, "密码错误"}
	ErrCaptchaError        = ErrorCode{10003, "验证码错误"}
	ErrUserNotExist        = ErrorCode{10004, "用户不存在"}
	ErrTokenInvalid        = ErrorCode{10005, "Token失效"}
	ErrRepeatOperation     = ErrorCode{10007, "重复操作"}
	ErrEmailSendWait       = ErrorCode{10008, "邮箱已经发送 稍后再试"}
	ErrPasswordOrUserError = ErrorCode{10009, "密码错误或用户不存在"}
	ErrCaptchaSendError    = ErrorCode{20001, "验证码发送失败"}
	ErrRegisterError       = ErrorCode{20002, "注册失败"}
	ErrOperateError        = ErrorCode{20003, "操作失败"}
	ErrUsernameExist       = ErrorCode{30001, "用户名已存在"}
	ErrEmailExist          = ErrorCode{30002, "邮箱已存在"}
	ErrParamInvalid        = ErrorCode{60001, "参数无效"}
	ErrParamBlank          = ErrorCode{60002, "参数为空"}
)

func ErrorEnum(c interface{ JSON(int, any) }, err ErrorCode) {
	c.JSON(200, Body{Code: err.Code, Msg: err.Msg, Data: nil})
}
