package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProposalAttachment struct {
	AttachmentID uuid.UUID `json:"attachment_id" gorm:"type:varchar(36);primaryKey"`
	ProposalID   uuid.UUID `json:"proposal_id" gorm:"type:varchar(36);not null;index"`
	FilePath     string    `json:"file_path" gorm:"type:text;not null"`
	OriginalName string    `json:"original_name" gorm:"type:varchar(255);not null"`
	MimeType     string    `json:"mime_type" gorm:"type:varchar(100);not null"`
	FileSize     int64     `json:"file_size" gorm:"type:bigint;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp"`
}
