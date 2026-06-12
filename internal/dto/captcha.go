package dto

type EmailCaptchaRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type ResetPasswordTokenResponse struct {
	ResetToken string `json:"resetToken"`
}
