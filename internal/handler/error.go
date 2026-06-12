package handler

import (
	"errors"

	"onepractice-golang/internal/response"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCaptchaInvalid):
		response.ErrorEnum(c, response.ErrCaptchaError)
	case errors.Is(err, service.ErrEmailSendWait):
		response.ErrorEnum(c, response.ErrEmailSendWait)
	case errors.Is(err, service.ErrUsernameExists):
		response.ErrorEnum(c, response.ErrUsernameExist)
	case errors.Is(err, service.ErrEmailExists):
		response.ErrorEnum(c, response.ErrEmailExist)
	case errors.Is(err, service.ErrPasswordOrUserError):
		response.ErrorEnum(c, response.ErrPasswordOrUserError)
	case errors.Is(err, service.ErrTokenInvalid):
		response.ErrorEnum(c, response.ErrTokenInvalid)
	case errors.Is(err, service.ErrInvalidParam):
		response.ErrorEnum(c, response.ErrParamInvalid)
	case errors.Is(err, service.ErrDatabaseDisabled):
		response.Error(c, 500, err.Error())
	default:
		response.Error(c, response.ErrOperateError.Code, err.Error())
	}
}
