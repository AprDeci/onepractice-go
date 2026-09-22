package dto

// 以下为 legacy（/api/*）线格式，字段沿用旧名以免打断弃用中的客户端；语义已改为昵称/邮箱。

type RegisterRequest struct {
	Nickname    string `json:"username" binding:"required,min=3,max=20"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	CaptchaCode string `json:"captchacode" binding:"required"`
	UserType    int    `json:"usertype"`
}

type LoginRequest struct {
	// legacy 线格式仍是 usernameorEmail，但只接受邮箱。
	Email    string `json:"usernameorEmail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email      string `json:"email" binding:"required,email"`
	ResetToken string `json:"resetToken" binding:"required"`
	Password   string `json:"password" binding:"required"`
}
