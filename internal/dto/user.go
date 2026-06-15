package dto

type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=20"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	CaptchaCode string `json:"captchacode" binding:"required"`
	UserType    int    `json:"usertype"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"usernameorEmail" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type UserInfoResponse struct {
	Username string `json:"username"`
	UserType int    `json:"userType"`
	Email    string `json:"email"`
}

type ResetPasswordRequest struct {
	Email      string `json:"email" binding:"required,email"`
	ResetToken string `json:"resetToken" binding:"required"`
	Password   string `json:"password" binding:"required"`
}
