package dto

type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=20"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	CaptchaCode string `json:"captchacode" binding:"required"`
	UserType    int    `json:"usertype"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameorEmail" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email      string `json:"email" binding:"required,email"`
	ResetToken string `json:"resetToken" binding:"required"`
	Password   string `json:"password" binding:"required"`
}
