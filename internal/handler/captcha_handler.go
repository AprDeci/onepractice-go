package handler

import (
	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/response"
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
	if err := h.service.SendEmailCaptcha(c.Query("email")); err != nil {
		writeError(c, err)
		return
	}
	response.SuccessNoData(c)
}

// VerifyEmail 校验邮箱验证码。
// @Summary 校验邮箱验证码
// @Description 校验邮箱验证码，成功后返回重置密码凭证。
// @Tags captcha
// @Accept json
// @Produce json
// @Param request body dto.EmailCaptchaRequest true "邮箱验证码参数"
// @Success 200 {object} response.Body
// @Router /api/captcha/email/verify [post]
func (h *CaptchaHandler) VerifyEmail(c *gin.Context) {
	var req dto.EmailCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorEnum(c, response.ErrParamInvalid)
		return
	}

	token, err := h.service.VerifyResetPassword(req.Email, req.Code)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, dto.ResetPasswordTokenResponse{ResetToken: token})
}
