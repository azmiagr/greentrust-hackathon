package model

import "github.com/google/uuid"

type GetOtp struct {
	OtpID  uuid.UUID `json:"otp_id"`
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

type VerifyOTPParam struct {
	Code string `json:"code" binding:"required,len=6"`
}

type VerifyOTPResponse struct {
	SessionToken string `json:"session_token,omitempty"`
	Message      string `json:"message"`
}
