package entity

import (
	"time"

	"github.com/google/uuid"
)

type Proposal struct {
	ProposalID        uuid.UUID  `json:"proposal_id" gorm:"type:varchar(36);primaryKey"`
	SenderUserID      uuid.UUID  `json:"sender_user_id" gorm:"type:varchar(36);not null;index"`
	ReceiverUserID    uuid.UUID  `json:"receiver_user_id" gorm:"type:varchar(36);not null;index"`
	SenderRole        string     `json:"sender_role" gorm:"type:enum('umkm','investor');not null;index"`
	ReceiverRole      string     `json:"receiver_role" gorm:"type:enum('umkm','investor');not null;index"`
	SenderProfileID   uuid.UUID  `json:"sender_profile_id" gorm:"type:varchar(36);not null;index"`
	ReceiverProfileID uuid.UUID  `json:"receiver_profile_id" gorm:"type:varchar(36);not null;index"`
	ProposalType      string     `json:"proposal_type" gorm:"type:enum('funding','supply');not null"`
	Title             string     `json:"title" gorm:"type:varchar(200);not null"`
	Amount            int64      `json:"amount" gorm:"type:bigint;not null"`
	TenorMonths       int        `json:"tenor_months" gorm:"type:int"`
	Scheme            string     `json:"scheme" gorm:"type:varchar(100)"`
	Message           string     `json:"message" gorm:"type:text"`
	Status            string     `json:"status" gorm:"type:enum('draft','sent','accepted','rejected','withdrawn');default:'draft';index"`
	SentAt            *time.Time `json:"sent_at" gorm:"type:timestamp;null"`
	AcceptedAt        *time.Time `json:"accepted_at" gorm:"type:timestamp;null"`
	RejectedAt        *time.Time `json:"rejected_at" gorm:"type:timestamp;null"`
	CreatedAt         time.Time  `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"type:timestamp"`

	Sender      User                 `json:"sender" gorm:"foreignKey:SenderUserID;references:UserID;constraint:onDelete:CASCADE"`
	Receiver    User                 `json:"receiver" gorm:"foreignKey:ReceiverUserID;references:UserID;constraint:onDelete:CASCADE"`
	Attachments []ProposalAttachment `json:"attachments" gorm:"foreignKey:ProposalID;constraint:onDelete:CASCADE"`
}
