package entity

import (
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	OtpID     uuid.UUID `json:"otp_id" gorm:"type:varchar(36);primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null"`
	Code      string    `json:"code" gorm:"type:varchar(6);not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp"`
}
