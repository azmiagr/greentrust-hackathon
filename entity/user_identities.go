package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserIdentity struct {
	IdentityID   uuid.UUID `json:"identity_id" gorm:"type:varchar(36);primaryKey"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	KTPFilePath  string    `json:"ktp_file_path" gorm:"type:text;not null"`
	FirstName    string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName     string    `json:"last_name" gorm:"type:varchar(100);not null"`
	BirthPlace   string    `json:"birth_place" gorm:"type:varchar(200);not null"`
	BirthDate    time.Time `json:"birth_date" gorm:"type:date;not null"`
	NIK          string    `json:"nik" gorm:"type:varchar(16);uniqueIndex;not null"`
	Address      string    `json:"address" gorm:"type:text;not null"`
	Province     string    `json:"province" gorm:"type:varchar(100);not null"`
	City         string    `json:"city" gorm:"type:varchar(100);not null"`
	PhoneNumber  string    `json:"phone_number" gorm:"type:varchar(20);not null"`
	EmailContact string    `json:"email_contact" gorm:"type:varchar(200);not null"`
	IsConfirmed  bool      `json:"is_confirmed" gorm:"type:boolean"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamp"`
}
