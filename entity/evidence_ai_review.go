package entity

import (
	"time"

	"github.com/google/uuid"
)

type EvidenceAIReview struct {
	ReviewID               uuid.UUID  `json:"review_id" gorm:"type:varchar(36);primaryKey"`
	EvidenceID             uuid.UUID  `json:"evidence_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	ProfileID              uuid.UUID  `json:"profile_id" gorm:"type:varchar(36);not null;index"`
	AISummary              string     `json:"ai_summary" gorm:"type:text"`
	AIConfidence           float64    `json:"ai_confidence" gorm:"type:decimal(5,2);default:0"`
	SuggestedCategoryID    string     `json:"suggested_category_id" gorm:"type:varchar(10);index"`
	SuggestedRequirementID *string    `json:"suggested_requirement_id" gorm:"type:varchar(36);index"`
	FinalRequirementID     *string    `json:"final_requirement_id" gorm:"type:varchar(36);index"`
	ReviewStatus           string     `json:"review_status" gorm:"type:enum('pending','approved','corrected','rejected');default:'pending'"`
	ReviewedByUserID       *uuid.UUID `json:"reviewed_by_user_id" gorm:"type:varchar(36);null"`
	ReviewedAt             *time.Time `json:"reviewed_at" gorm:"type:timestamp;null"`
	CreatedAt              time.Time  `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt              time.Time  `json:"updated_at" gorm:"type:timestamp"`
}
