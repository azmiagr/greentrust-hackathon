package model

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type EvidenceSummaryResponse struct {
	GRSScore          float64                    `json:"grs_score"`
	PassportThreshold float64                    `json:"passport_threshold"`
	PassportStatus    string                     `json:"passport_status"`
	Categories        []EvidenceCategoryProgress `json:"categories"`
}

type EvidenceCategoryProgress struct {
	CategoryID     string  `json:"category_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Weight         float64 `json:"weight"`
	RequiredCount  int     `json:"required_count"`
	FulfilledCount int     `json:"fulfilled_count"`
	Score          float64 `json:"score"`
	Status         string  `json:"status"` // empty, partial, complete
}

type EvidenceCategoryDetailResponse struct {
	Category     EvidenceCategoryProgress   `json:"category"`
	Requirements []EvidenceRequirementItem  `json:"requirements"`
	Documents    []EvidenceDocumentItem     `json:"documents"`
	NextPriority []EvidenceCategoryProgress `json:"next_priority"`
}

type EvidenceRequirementItem struct {
	RequirementID string                `json:"requirement_id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	IsRequired    bool                  `json:"is_required"`
	Document      *EvidenceDocumentItem `json:"document,omitempty"`
}

type EvidenceDocumentItem struct {
	EvidenceID    uuid.UUID `json:"evidence_id"`
	RequirementID *string   `json:"requirement_id,omitempty"`
	FileName      string    `json:"file_name"`
	FilePath      string    `json:"file_path"`
	MimeType      string    `json:"mime_type"`
	FileSize      int64     `json:"file_size"`
	Status        string    `json:"status"`
	AiConfidence  float64   `json:"ai_confidence"`
	CreatedAt     time.Time `json:"created_at"`
}

type UploadEvidenceDocumentParam struct {
	CategoryID    string
	RequirementID string
	File          *multipart.FileHeader
	AIConfidence  float64
}

type EvidenceDocumentResponse struct {
	EvidenceID       uuid.UUID `json:"evidence_id"`
	ProfileID        uuid.UUID `json:"profile_id"`
	RequirementID    *string   `json:"requirement_id,omitempty"`
	FileName         string    `json:"file_name"`
	FilePath         string    `json:"file_path"`
	FileHash         string    `json:"file_hash"`
	MimeType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	AiConfidence     float64   `json:"ai_confidence"`
	Status           string    `json:"status"`
	BlockchainTxHash string    `json:"blockchain_tx_hash,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type EvidenceCategoryResponse struct {
	CategoryID   string                        `json:"category_id"`
	Code         string                        `json:"code"`
	Name         string                        `json:"name"`
	Weight       float64                       `json:"weight"`
	SortOrder    int                           `json:"sort_order"`
	Requirements []EvidenceRequirementResponse `json:"requirements"`
}

type EvidenceRequirementResponse struct {
	RequirementID string `json:"requirement_id"`
	CategoryID    string `json:"category_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsRequired    bool   `json:"is_required"`
	SortOrder     int    `json:"sort_order"`
}
