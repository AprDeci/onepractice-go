package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

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
	ErrUsernameExists      = errors.New("用户名已存在")
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

func (s *UserService) Register(req dto.RegisterRequest) (dto.RegisterResponse, error) {
	if s.db == nil {
		return dto.RegisterResponse{}, ErrDatabaseDisabled
	}

	username := strings.TrimSpace(req.Username)
	email := normalizeEmail(req.Email)
	if username == "" || email == "" || req.Password == "" {
		return dto.RegisterResponse{}, ErrInvalidParam
	}

	if err := s.captcha.VerifyRegister(email, req.CaptchaCode); err != nil {
		return dto.RegisterResponse{}, err
	}

	if exists, err := s.usernameExists(username); err != nil {
		return dto.RegisterResponse{}, err
	} else if exists {
		return dto.RegisterResponse{}, ErrUsernameExists
	}
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
		Username: username,
		Password: passwordHash,
		Email:    storedEmail,
		UserType: req.UserType,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return dto.RegisterResponse{}, err
	}

	return dto.RegisterResponse{Username: username, Email: email}, nil
}

func (s *UserService) Login(req dto.LoginRequest) (dto.LoginResponse, error) {
	if s.db == nil {
		return dto.LoginResponse{}, ErrDatabaseDisabled
	}

	account := strings.TrimSpace(req.UsernameOrEmail)
	var user model.User
	query := s.db
	if strings.Contains(account, "@") {
		email := normalizeEmail(account)
		encEmail, err := utils.LegacyAESEncrypt(email)
		if err != nil {
			return dto.LoginResponse{}, err
		}
		query = query.Where("email in ?", []string{email, encEmail})
	} else {
		query = query.Where("username = ?", account)
	}

	if err := query.First(&user).Error; err != nil {
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

	return dto.LoginResponse{ID: user.ID, Username: user.Username, Email: decryptEmail(user.Email), Token: token}, nil
}

func (s *UserService) Info(userID int64) (dto.UserInfoResponse, error) {
	if s.db == nil {
		return dto.UserInfoResponse{}, ErrDatabaseDisabled
	}

	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return dto.UserInfoResponse{}, err
	}
	return dto.UserInfoResponse{Username: user.Username, UserType: user.UserType, Email: decryptEmail(user.Email)}, nil
}

func (s *UserService) ResetPassword(req dto.ResetPasswordRequest) error {
	if s.db == nil {
		return ErrDatabaseDisabled
	}

	email := normalizeEmail(req.Email)
	if s.captcha != nil && s.captcha.UsesRedis() {
		if err := s.captcha.ConsumeResetToken(email, req.ResetToken); err != nil {
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
		return s.db.Model(&model.User{}).Where("email in ?", []string{email, encEmail}).Update("password", passwordHash).Error
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var token model.PasswordResetToken
		if err := tx.Where("email = ? and token = ? and consumed_at is null", email, req.ResetToken).
			Order("created_at desc").First(&token).Error; err != nil {
			return ErrInvalidParam
		}
		if token.ExpiresAt.Before(time.Now()) {
			return ErrInvalidParam
		}

		passwordHash, err := utils.HashPassword(req.Password)
		if err != nil {
			return err
		}
		encEmail, err := utils.LegacyAESEncrypt(email)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.User{}).Where("email in ?", []string{email, encEmail}).Update("password", passwordHash).Error; err != nil {
			return err
		}
		n := time.Now()
		return tx.Model(&token).Update("consumed_at", &n).Error
	})
}

func LoginIDFromToken(token string) (int64, error) {
	loginID, err := sagin.GetLoginID(token)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(loginID, 10, 64)
}

func (s *UserService) usernameExists(username string) (bool, error) {
	var count int64
	err := s.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
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
