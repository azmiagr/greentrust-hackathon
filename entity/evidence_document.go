package entity

import (
	"time"

	"github.com/google/uuid"
)

type EvidenceDocument struct {
	EvidenceID       uuid.UUID `json:"evidence_id" gorm:"type:varchar(36);primaryKey"`
	ProfileID        uuid.UUID `json:"profile_id" gorm:"type:varchar(36);not null;index"`
	EvidenceCategory string    `json:"evidence_category" gorm:"type:varchar(100);"`
	FilePath         string    `json:"file_path" gorm:"type:text;not null"`
	FileHash         string    `json:"file_hash" gorm:"type:varchar(255);not null"`
	AiConfidence     float64   `json:"ai_confidence" gorm:"type:decimal(3,2);"`
	Status           string    `json:"status" gorm:"type:enum('uploading', 'processing', 'classified', 'reviewed', 'on_chain', 'blockchain_failed');"`
	BlockchainTxHash string    `json:"blockchain_tx_hash" gorm:"type:varchar(255);null"`
	CreatedAt        time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"type:timestamp"`
}
