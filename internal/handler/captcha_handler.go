package handler

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type CaptchaHandler struct {
	service *service.CaptchaService
}

func NewCaptchaHandler(service *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{service: service}
}

// Email 发送邮箱验证码。
// @Summary 发送邮箱验证码
// @Description 向指定邮箱发送验证码。
// @Tags captcha
// @Produce json
// @Param email query string true "邮箱"
// @Success 200 {object} response.Body
// @Router /api/captcha/email [get]
func (h *CaptchaHandler) Email(c *gin.Context) {
	if err := h.service.SendEmailCaptcha(c.Request.Context(), c.Query("email"), c.DefaultQuery("purpose", service.CaptchaPurposeRegister)); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrEmailExists):
			response.Error(c, apperror.New(apperror.CodeConflict, "资源已存在"))
		case errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		case errors.Is(err, service.ErrEmailSendWait):
			response.Error(c, apperror.New(apperror.CodeConflict, "请稍后重试"))
		default:
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		}
		return
	}
	response.Success(c, nil)
}

// VerifyEmail 校验邮箱验证码。
// @Summary 校验邮箱验证码
// @Description 校验邮箱验证码，成功后返回重置密码凭证。
// @Tags captcha
// @Accept json
// @Produce json
// @Param request body dto.EmailCaptchaRequest true "邮箱验证码参数"
// @Success 200 {object} response.Body{data=dto.ResetPasswordTokenResponse}
// @Router /api/captcha/email/verify [post]
func (h *CaptchaHandler) VerifyEmail(c *gin.Context) {
	var req dto.EmailCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	token, err := h.service.VerifyResetPassword(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrCaptchaInvalid):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "验证码错误"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		}
		return
	}
	response.Success(c, dto.ResetPasswordTokenResponse{ResetToken: token})
}
