package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID    uuid.UUID `json:"user_id" gorm:"type:varchar(36);primaryKey"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:varchar(36);not null"`
	Email     string    `json:"email" gorm:"type:varchar(200);not null;uniqueIndex"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:enum('active','inactive')"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp"`

	OTPs              []OTP           `json:"otps" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	UserIdentity      UserIdentity    `json:"user_identity" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	InvestorProfile   InvestorProfile `json:"investor_profile" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	UMKMProfile       UMKMProfile     `json:"umkm_profile" gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
	SentProposals     []Proposal      `json:"sent_proposals" gorm:"foreignKey:SenderUserID;constraint:onDelete:CASCADE"`
	ReceivedProposals []Proposal      `json:"received_proposals" gorm:"foreignKey:ReceiverUserID;constraint:onDelete:CASCADE"`
}
