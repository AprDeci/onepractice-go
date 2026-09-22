package apiv1

type RegisterRequest struct {
	Nickname    string `json:"nickname" binding:"required,min=3,max=20"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
	UserType    int    `json:"userType"`
}

// LoginRequest 只接受邮箱登录；昵称不唯一，不能作为登录凭证。
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email      string `json:"email" binding:"required,email"`
	ResetToken string `json:"resetToken" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type EmailVerificationRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Purpose string `json:"purpose"`
}

type EmailVerificationVerifyRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Code    string `json:"code" binding:"required"`
	Purpose string `json:"purpose"`
}
