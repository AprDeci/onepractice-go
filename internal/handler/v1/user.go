package v1

import (
	"errors"

	"onepractice-golang/internal/common/apperror"
	"onepractice-golang/internal/common/response"
	dto "onepractice-golang/internal/dto"
	dtoV1 "onepractice-golang/internal/dto/v1"
	"onepractice-golang/internal/service"

	"github.com/gin-gonic/gin"
	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

type UserHandler struct{ service *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{service: svc} }

// Register 用户注册。
// @Summary 用户注册
// @Description 使用用户名、密码、邮箱和邮箱验证码注册用户。
// @Tags user
// @Accept json
// @Produce json
// @Param request body apiv1.RegisterRequest true "注册参数"
// @Success 201 {object} response.Body
// @Router /api/v1/auth/registrations [post]
func (h *UserHandler) Register(c *gin.Context) {
	var input dtoV1.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	result, err := h.service.Register(c.Request.Context(), dto.RegisterRequest{
		Username: input.Username, Password: input.Password, Email: input.Email,
		CaptchaCode: input.CaptchaCode, UserType: input.UserType,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrUsernameExists), errors.Is(err, service.ErrEmailExists):
			response.Error(c, apperror.New(apperror.CodeConflict, "资源已存在"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		}
		return
	}
	response.Created(c, dtoV1.RegisterResponse{Username: result.Username, Email: result.Email})
}

// Login 用户登录。
// @Summary 用户登录
// @Description 使用用户名或邮箱登录。
// @Tags user
// @Accept json
// @Produce json
// @Param request body apiv1.LoginRequest true "登录参数"
// @Success 200 {object} response.Body
// @Router /api/v1/auth/sessions [post]
func (h *UserHandler) Login(c *gin.Context) {
	var input dtoV1.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	result, err := h.service.Login(dto.LoginRequest{UsernameOrEmail: input.UsernameOrEmail, Password: input.Password})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrPasswordOrUserError):
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "用户名或密码错误"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		}
		return
	}
	response.Success(c, dtoV1.LoginResponse{ID: result.ID, Username: result.Username, Email: result.Email, Token: result.Token})
}

// Logout 用户登出。
// @Summary 用户登出
// @Description 使当前 token 失效。
// @Tags user
// @Security ApiKeyAuth
// @Success 204
// @Router /api/v1/auth/sessions/current [delete]
func (h *UserHandler) Logout(c *gin.Context) {
	token := sagin.GetTokenFromCtx(c)
	if token == "" {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		return
	}
	if err := sagin.LogoutByToken(token); err != nil {
		response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		return
	}
	response.NoContent(c)
}

// Info 获取当前用户信息。
// @Summary 获取当前用户信息
// @Description 根据 token 获取当前登录用户信息。
// @Tags user
// @Security ApiKeyAuth
// @Success 200 {object} response.Body
// @Router /api/v1/users/me [get]
func (h *UserHandler) Info(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Info(userID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.Error(c, apperror.New(apperror.CodeNotFound, "资源不存在"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		}
		return
	}
	response.Success(c, dtoV1.UserInfoResponse{Username: result.Username, UserType: result.UserType, Email: result.Email})
}

// ResetPassword 重置密码。
// @Summary 重置密码
// @Description 使用邮箱和重置凭证重置密码。
// @Tags user
// @Accept json
// @Produce json
// @Param request body apiv1.ResetPasswordRequest true "重置密码参数"
// @Success 200 {object} response.Body
// @Router /api/v1/auth/password-resets [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var input dtoV1.ResetPasswordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		return
	}
	err := h.service.ResetPassword(c.Request.Context(), dto.ResetPasswordRequest{Email: input.Email, ResetToken: input.ResetToken, Password: input.Password})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParam), errors.Is(err, service.ErrCaptchaInvalid):
			response.Error(c, apperror.New(apperror.CodeInvalidArgument, "参数无效"))
		case errors.Is(err, service.ErrTokenInvalid):
			response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		case errors.Is(err, service.ErrDatabaseDisabled), errors.Is(err, service.ErrRedisDisabled):
			response.Error(c, apperror.New(apperror.CodeServiceUnavailable, "依赖服务不可用"))
		default:
			response.Error(c, apperror.Wrap(apperror.CodeInternal, "系统异常", err))
		}
		return
	}
	response.Success(c, nil)
}

func currentUserID(c *gin.Context) (int64, bool) {
	token := sagin.GetTokenFromCtx(c)
	if token == "" {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		return 0, false
	}
	id, err := service.LoginIDFromToken(token)
	if err != nil {
		response.Error(c, apperror.New(apperror.CodeUnauthorized, "Token失效"))
		return 0, false
	}
	return id, true
}
