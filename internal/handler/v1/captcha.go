package v1

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
)

type CaptchaHandler struct{ service *service.CaptchaService }

func NewCaptchaHandler(svc *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{service: svc}
}

// SendEmail 发送邮箱验证码。
// @Summary 发送邮箱验证码
// @Description 向指定邮箱发送验证码。
// @Tags captcha
// @Accept json
// @Produce json
// @Param request body apiv1.EmailVerificationRequest true "邮箱验证参数"
// @Success 200 {object} response.Body
// @Router /api/v1/auth/email-verifications [post]
func (h *CaptchaHandler) SendEmail(c *gin.Context) {
	var req dtoV1.EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	if err := h.service.SendEmailCaptcha(c.Request.Context(), req.Email, req.Purpose); err != nil {
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
// @Param request body apiv1.EmailVerificationVerifyRequest true "邮箱验证码参数"
// @Success 200 {object} response.Body{data=apiv1.ResetTokenResponse}
// @Router /api/v1/auth/email-verifications/verification [post]
func (h *CaptchaHandler) VerifyEmail(c *gin.Context) {
	var req dtoV1.EmailVerificationVerifyRequest
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
	response.Success(c, dtoV1.ResetTokenResponse{ResetToken: token})
}
