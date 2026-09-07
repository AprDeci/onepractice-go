package handler

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Register 用户注册。
// @Summary 用户注册
// @Description 使用用户名、密码、邮箱和邮箱验证码注册用户。
// @Tags user
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册参数"
// @Success 200 {object} response.Body
// @Router /api/user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	result, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrUsernameExists), errors.Is(err, service.ErrEmailExists):
			response.Error(c, apperror.New(apperror.CodeConflict, "资源已存在"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.New(apperror.CodeInternal, "系统异常"))
		}
		return
	}
	response.Success(c, result)
}

// Login 用户登录。
// @Summary 用户登录
// @Description 使用用户名或邮箱登录，成功后返回 sa-token-go token。兼容旧 AES 密码，成功后升级为 bcrypt。
// @Tags user
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录参数"
// @Success 200 {object} response.Body
// @Router /api/user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	result, err := h.service.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrPasswordOrUserError):
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "用户名或密码错误"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.New(apperror.CodeInternal, "系统异常"))
		}
		return
	}
	response.Success(c, result)
}

// Info 获取当前用户信息。
// @Summary 获取当前用户信息
// @Description 根据 token 获取当前登录用户信息。
// @Tags user
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Body
// @Router /api/user/info [get]
func (h *UserHandler) Info(c *gin.Context) {
	userID, ok := loginID(c)
	if !ok {
		return
	}

	info, err := h.service.Info(userID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.Error(c, apperror.New(apperror.CodeNotFound, "资源不存在"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.New(apperror.CodeInternal, "系统异常"))
		}
		return
	}
	response.Success(c, info)
}

// Logout 用户登出。
// @Summary 用户登出
// @Description 使当前 token 失效。
// @Tags user
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Body
// @Router /api/user/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	token := sagin.GetTokenFromCtx(c)
	if token == "" {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		return
	}
	if err := sagin.LogoutByToken(token); err != nil {
		response.Error(c, apperror.New(apperror.CodeInternal, "系统异常"))
		return
	}
	response.Success(c, nil)
}

// ResetPassword 重置密码。
// @Summary 重置密码
// @Description 使用邮箱和重置凭证重置密码。
// @Tags user
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "重置密码参数"
// @Success 200 {object} response.Body
// @Router /api/user/resetpassword [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam), errors.Is(err, service.ErrCaptchaInvalid):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrTokenInvalid):
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.New(apperror.CodeInternal, "系统异常"))
		}
		return
	}
	response.Success(c, nil)
}

func loginID(c *gin.Context) (int64, bool) {
	token := sagin.GetTokenFromCtx(c)
	if token == "" {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		return 0, false
	}
	loginID, err := service.LoginIDFromToken(token)
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		return 0, false
	}
	return loginID, true
}
