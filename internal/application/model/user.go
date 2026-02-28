package model

type GetUserRequest struct {
	Page   int    `json:"page" validate:"omitempty,number,max=50" example:"1"`
	Limit  int    `json:"limit" validate:"omitempty,number,max=50" example:"10"`
	Search string `json:"search" validate:"omitempty,max=50" example:"example"`
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,max=50" example:"fake name"`
	Email    string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	Password string `json:"password" validate:"required,min=8,max=20,password" example:"password1"`
	Role     string `json:"role" validate:"required,oneof=user admin,max=50" example:"user"`
}

type CreateUserResponse struct {
	ID              string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name            string `json:"name" example:"fake name"`
	Email           string `json:"email" example:"fake@example.com"`
	Role            string `json:"role" example:"user"`
	IsEmailVerified bool   `json:"is_email_verified" example:"false"`
}

type UpdatePassOrVerifyRequest struct {
	Password      string `json:"password" validate:"omitempty,min=8,max=20,password" example:"password1"`
	VerifiedEmail bool   `json:"verified_email" validate:"omitempty,boolean" example:"false"`
}

type UpdateUserRequest struct {
	UserID   string `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name     string `json:"name" validate:"omitempty,max=50" example:"fake name"`
	Email    string `json:"email" validate:"omitempty,email,max=50" example:"fake@example.com"`
	Password string `json:"password" validate:"omitempty,min=8,max=20,password" example:"password1"`
}

type UpdateUserResponse struct {
	ID              string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name            string `json:"name" example:"fake name"`
	Email           string `json:"email" example:"fake@example.com"`
	Role            string `json:"role" example:"user"`
	IsEmailVerified bool   `json:"is_email_verified" example:"false"`
}

type CreateGoogleUserRequest struct {
	Name          string `json:"name" validate:"required,max=50" example:"fake name"`
	Email         string `json:"email" validate:"required,email,max=50" example:"fake@example.com"`
	VerifiedEmail bool   `json:"verified_email" validate:"omitempty,boolean" example:"false"`
}

type GetUserResponse struct {
	ID              string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name            string `json:"name" example:"fake name"`
	Email           string `json:"email" example:"fake@example.com"`
	Role            string `json:"role" example:"user"`
	IsEmailVerified bool   `json:"is_email_verified" example:"false"`
}
