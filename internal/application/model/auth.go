package model

import "time"

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,max=50" example:"fake name"`
	Email    string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password string `json:"password" validate:"required,min=8,max=20,strong-password" example:"password1"`
}

type RegisterResponse struct {
	ID                    string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name                  string    `json:"name" example:"fake name"`
	Email                 string    `json:"email" example:"fake@example.com"`
	AccessToken           string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" example:"2024-10-07T11:56:46.618180553Z"`
	RefreshToken          string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" example:"2024-10-07T11:56:46.618180553Z"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password string `json:"password" validate:"required,min=8,max=20,strong-password" example:"password1"`
}

type LoginResponse struct {
	AccessToken           string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" example:"2024-10-07T11:56:46.618180553Z"`
	RefreshToken          string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" example:"2024-10-07T11:56:46.618180553Z"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=255" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,max=255" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
}

type RefreshTokenResponse struct {
	AccessToken string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
	ExpiresAt   time.Time `json:"expires_at" example:"2024-10-07T11:56:46.618180553Z"`
}

type ForgotPasswordRequest struct {
	Email    string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password string `json:"password" validate:"required,min=8,max=20,password" example:"password1"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,max=255" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
	Password string `json:"password" validate:"required,min=8,max=20,password" example:"password1"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required,max=255" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1ZWJhYzUzNDk1NGI1NDEzOTgwNmMxMTIiLCJpYXQiOjE1ODkyOTg0ODQsImV4cCI6MTU4OTMwMDI4NH0.m1U63blB0MLej_WfB7yC2FTMnCziif9X8yzwDEfJXAg"`
}

type VerifyEmailResponse struct {
	Code    int    `json:"code" example:"200"`
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Email verified successfully"`
}

type GoogleLoginRequest struct {
	Code string `json:"code" validate:"required,max=255"`
}
