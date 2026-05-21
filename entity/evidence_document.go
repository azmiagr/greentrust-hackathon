package entity

import (
	"time"

	"github.com/google/uuid"
)

type EvidenceDocument struct {
	EvidenceID       uuid.UUID `json:"evidence_id" gorm:"type:varchar(36);primaryKey"`
	ProfileID        uuid.UUID `json:"profile_id" gorm:"type:varchar(36);not null;index"`
	RequirementID    *string   `json:"requirement_id" gorm:"type:varchar(36);index"`
	FilePath         string    `json:"file_path" gorm:"type:text;not null"`
	FileHash         string    `json:"file_hash" gorm:"type:varchar(255);not null"`
	AiConfidence     float64   `json:"ai_confidence" gorm:"type:decimal(3,2);"`
	Status           string    `json:"status" gorm:"type:enum('uploaded','processing','classified','reviewed','rejected','on_chain','blockchain_failed');default:'uploaded'"`
	BlockchainTxHash string    `json:"blockchain_tx_hash" gorm:"type:varchar(255);null"`
	OriginalName     string    `json:"original_name" gorm:"type:varchar(255)"`
	MimeType         string    `json:"mime_type" gorm:"type:varchar(100)"`
	FileSize         int64     `json:"file_size" gorm:"type:bigint"`
	CreatedAt        time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"type:timestamp"`

	EvidenceAIReviews []EvidenceAIReview `json:"evidence_ai_review" gorm:"foreignKey:EvidenceID;constraint:onDelete:CASCADE"`
}
