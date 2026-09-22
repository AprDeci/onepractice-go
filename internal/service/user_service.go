package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"onepractice-golang/internal/dto"
	"onepractice-golang/internal/model"
	"onepractice-golang/internal/utils"

	sagin "github.com/sa-tokens/sa-token-go/integrations/gin"
	"gorm.io/gorm"
)

var (
	ErrInvalidParam        = errors.New("参数无效")
	ErrCaptchaInvalid      = errors.New("验证码错误")
	ErrEmailSendWait       = errors.New("邮箱已经发送 稍后再试")
	ErrEmailExists         = errors.New("邮箱已存在")
	ErrPasswordOrUserError = errors.New("密码错误或用户不存在")
	ErrTokenInvalid        = errors.New("Token失效")
)

type UserService struct {
	db      *gorm.DB
	captcha *CaptchaService
}

func NewUserService(db *gorm.DB, captcha *CaptchaService) *UserService {
	return &UserService{db: db, captcha: captcha}
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) (dto.RegisterResponse, error) {
	if s.db == nil {
		return dto.RegisterResponse{}, ErrDatabaseDisabled
	}

	nickname := strings.TrimSpace(req.Nickname)
	email := normalizeEmail(req.Email)
	if nickname == "" || email == "" || req.Password == "" {
		return dto.RegisterResponse{}, ErrInvalidParam
	}

	if err := s.captcha.VerifyRegister(ctx, email, req.CaptchaCode); err != nil {
		return dto.RegisterResponse{}, err
	}

	// 昵称只是展示名，可以重复；邮箱才是唯一标识。
	if exists, err := s.emailExists(email); err != nil {
		return dto.RegisterResponse{}, err
	} else if exists {
		return dto.RegisterResponse{}, ErrEmailExists
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return dto.RegisterResponse{}, err
	}
	storedEmail, err := utils.LegacyAESEncrypt(email)
	if err != nil {
		return dto.RegisterResponse{}, err
	}

	user := model.User{
		Nickname: nickname,
		Password: passwordHash,
		Email:    storedEmail,
		UserType: req.UserType,
	}
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return dto.RegisterResponse{}, err
	}

	return dto.RegisterResponse{Nickname: nickname, Email: email}, nil
}

func (s *UserService) Login(req dto.LoginRequest) (dto.LoginResponse, error) {
	if s.db == nil {
		return dto.LoginResponse{}, ErrDatabaseDisabled
	}

	account := normalizeEmail(req.Email)
	if account == "" {
		return dto.LoginResponse{}, ErrInvalidParam
	}

	encEmail, err := utils.LegacyAESEncrypt(account)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	var user model.User
	if err := s.db.Where("email in ?", []string{account, encEmail}).First(&user).Error; err != nil {
		return dto.LoginResponse{}, ErrPasswordOrUserError
	}

	matched, shouldUpgrade := utils.VerifyPassword(user.Password, req.Password)
	if !matched {
		return dto.LoginResponse{}, ErrPasswordOrUserError
	}
	if shouldUpgrade {
		if newHash, err := utils.HashPassword(req.Password); err == nil {
			_ = s.db.Model(&user).Update("password", newHash).Error
		}
	}

	token, err := sagin.Login(user.ID)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.LoginResponse{ID: user.ID, Nickname: user.Nickname, Email: decryptEmail(user.Email), Token: token}, nil
}

func (s *UserService) Info(userID int64) (dto.UserInfoResponse, error) {
	if s.db == nil {
		return dto.UserInfoResponse{}, ErrDatabaseDisabled
	}

	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return dto.UserInfoResponse{}, err
	}
	return dto.UserInfoResponse{Nickname: user.Nickname, UserType: user.UserType, Email: decryptEmail(user.Email)}, nil
}

func (s *UserService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	if s.db == nil {
		return ErrDatabaseDisabled
	}

	email := normalizeEmail(req.Email)
	if s.captcha == nil {
		return ErrRedisDisabled
	}

	if err := s.captcha.ConsumeResetToken(ctx, email, req.ResetToken); err != nil {
		return err
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	encEmail, err := utils.LegacyAESEncrypt(email)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&model.User{}).Where("email in ?", []string{email, encEmail}).Update("password", passwordHash).Error
}

func LoginIDFromToken(token string) (int64, error) {
	loginID, err := sagin.GetLoginID(token)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(loginID, 10, 64)
}

func (s *UserService) emailExists(email string) (bool, error) {
	encEmail, err := utils.LegacyAESEncrypt(email)
	if err != nil {
		return false, err
	}
	var count int64
	err = s.db.Model(&model.User{}).Where("email in ?", []string{email, encEmail}).Count(&count).Error
	return count > 0, err
}

func decryptEmail(stored string) string {
	plain, err := utils.LegacyAESDecrypt(stored)
	if err != nil {
		return stored
	}
	return plain
}
