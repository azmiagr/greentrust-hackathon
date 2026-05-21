package model

import (
	"time"

	"github.com/google/uuid"
)

type SubmitEvidenceAIReviewParam struct {
	EvidenceID             uuid.UUID `json:"-"`
	AISummary              string    `json:"ai_summary"`
	AIConfidence           float64   `json:"ai_confidence"`
	SuggestedCategoryID    string    `json:"suggested_category_id"`
	SuggestedRequirementID *string   `json:"suggested_requirement_id"`
}

type ReviewEvidenceAIParam struct {
	ReviewID           uuid.UUID `json:"-"`
	Action             string    `json:"action"` // approve, correct, reject
	FinalRequirementID *string   `json:"final_requirement_id,omitempty"`
}

type EvidenceAIReviewResponse struct {
	ReviewID               uuid.UUID             `json:"review_id"`
	EvidenceID             uuid.UUID             `json:"evidence_id"`
	ProfileID              uuid.UUID             `json:"profile_id"`
	AISummary              string                `json:"ai_summary"`
	AIConfidence           float64               `json:"ai_confidence"`
	SuggestedCategoryID    string                `json:"suggested_category_id"`
	SuggestedRequirementID *string               `json:"suggested_requirement_id,omitempty"`
	FinalRequirementID     *string               `json:"final_requirement_id,omitempty"`
	ReviewStatus           string                `json:"review_status"`
	ReviewedAt             *time.Time            `json:"reviewed_at,omitempty"`
	Document               *EvidenceDocumentItem `json:"document,omitempty"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}
