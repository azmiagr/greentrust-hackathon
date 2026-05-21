package model

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type SubmitUserIdentityParam struct {
	KTPFile      *multipart.FileHeader `form:"ktp_file"`
	FirstName    string                `form:"first_name" binding:"required"`
	LastName     string                `form:"last_name" binding:"required"`
	NIK          string                `form:"nik" binding:"required,len=16"`
	BirthPlace   string                `form:"birth_place" binding:"required"`
	BirthDate    string                `form:"birth_date" binding:"required"`
	Address      string                `form:"address" binding:"required"`
	Province     string                `form:"province" binding:"required"`
	City         string                `form:"city" binding:"required"`
	PhoneNumber  string                `form:"phone_number" binding:"required"`
	EmailContact string                `form:"email_contact" binding:"required,email"`
	IsConfirmed  bool                  `form:"is_confirmed"`
}

type GetUserIdentityParam struct {
	IdentityID uuid.UUID `json:"identity_id"`
	UserID     uuid.UUID `json:"user_id"`
	NIK        string    `json:"nik"`
}

type SubmitUserIdentityResponse struct {
	IdentityID   uuid.UUID `json:"identity_id"`
	SessionToken string    `json:"session_token"`
	Message      string    `json:"message"`
}
