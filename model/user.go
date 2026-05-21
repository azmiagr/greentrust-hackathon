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
}

type RegisterResponse struct {
	SessionToken string `json:"session_token"`
	Message      string `json:"message"`
}
