package model

import "github.com/google/uuid"

type GetUserParam struct {
	UserID uuid.UUID `json:"-"`
	Email  string    `json:"-"`
}

type RegisterUserParam struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
	Role            string `json:"role" binding:"required,oneof=umkm investor"`
}

type RegisterResponse struct {
	SessionToken string `json:"session_token"`
	Message      string `json:"message"`
}

type LoginUserParam struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
