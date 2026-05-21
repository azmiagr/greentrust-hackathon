package model

import "github.com/google/uuid"

type SubmitUserIdentityParam struct {
	KTPFilePath  string `json:"ktp_file_path" binding:"required"`
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	NIK          string `json:"nik" binding:"required,len=16"`
	BirthPlace   string `json:"birth_place" binding:"required"`
	BirthDate    string `json:"birth_date" binding:"required"`
	Address      string `json:"address" binding:"required"`
	Province     string `json:"province" binding:"required"`
	City         string `json:"city" binding:"required"`
	PhoneNumber  string `json:"phone_number" binding:"required"`
	EmailContact string `json:"email_contact" binding:"required,email"`
	IsConfirmed  bool   `json:"is_confirmed"`
}

type GetUserIdentityParam struct {
	IdentityID uuid.UUID `json:"identity_id"`
	UserID     uuid.UUID `json:"user_id"`
	NIK        string    `json:"nik"`
}

type SubmitUserIdentityResponse struct {
	IdentityID uuid.UUID `json:"identity_id"`
	Message    string    `json:"message"`
}
