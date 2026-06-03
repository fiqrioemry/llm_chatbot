package dto

import "time"

 
type RegisterRequest struct {
	Fullname string `json:"fullname" binding:"required,min=1,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword"     binding:"required,min=8,max=72"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8,max=72"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"required,eqfield=Password"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailRequest struct {
	Token string `form:"token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type TokenPair struct {
	AccessToken           string    `json:"accessToken"`
	RefreshToken          string    `json:"refreshToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

type LoginResponse struct {
	Tokens TokenPair   `json:"tokens"`
	User   UserSummary `json:"user"`
}

type UserSummary struct {
	ID       string  `json:"id"`
	Fullname string  `json:"fullname"`
	Email    string  `json:"email"`
	Avatar   *string `json:"avatar"`
}

type MeResponse struct {
	ID                   string     `json:"id"`
	Fullname             string     `json:"fullname"`
	Email                string     `json:"email"`
	Avatar               *string    `json:"avatar"`
	Role                 string     `json:"role"`
	LastLogin            *time.Time `json:"lastLogin"`
	JoinedAt             time.Time  `json:"joinedAt"`
	LastChangePasswordAt *time.Time `json:"lastChangePasswordAt"`
}

type SessionResponse struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"userAgent"`
	IPAddress string    `json:"ipAddress"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}