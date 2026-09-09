package apiv1

type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=20"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
	UserType    int    `json:"userType"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type ResetPasswordRequest struct {
	Email      string `json:"email" binding:"required,email"`
	ResetToken string `json:"resetToken" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type UserInfoResponse struct {
	Username string `json:"username"`
	UserType int    `json:"userType"`
	Email    string `json:"email"`
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

type ResetTokenResponse struct {
	ResetToken string `json:"resetToken"`
}
