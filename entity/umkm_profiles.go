package entity

import (
	"time"

	"github.com/google/uuid"
)

type UMKMProfile struct {
	ProfileID              uuid.UUID `json:"profile_id" gorm:"type:varchar(36);primaryKey"`
	UserID                 uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	SectorID               uuid.UUID `json:"sector_id" gorm:"type:varchar(36);not null;index"`
	BusinessName           string    `json:"business_name" gorm:"type:varchar(200);not null"`
	BusinessDescription    string    `json:"business_description" gorm:"type:text"`
	BusinessAddressLine    string    `json:"business_address_line" gorm:"type:text;not null"`
	BusinessProvince       string    `json:"business_province" gorm:"type:varchar(100);not null"`
	BusinessCity           string    `json:"business_city" gorm:"type:varchar(100);not null"`
	IsServiceBusiness      bool      `json:"is_service_business" gorm:"type:boolean"`
	WhatsappNumber         string    `json:"whatsapp_number" gorm:"type:varchar(20);not null"`
	ProfileCompletionScore float64   `json:"profile_completion_score" gorm:"type:decimal(5,2)"`
	CreatedAt              time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"type:timestamp"`

	GreenPassport     GreenPassport      `json:"green_passport" gorm:"foreignKey:ProfileID;constraint:onDelete:CASCADE"`
	LocationPhotos    []LocationPhoto    `json:"location_photos" gorm:"foreignKey:ProfileID;constraint:onDelete:CASCADE"`
	UMKMProducts      []UMKMProduct      `json:"umkm_products" gorm:"foreignKey:ProfileID;constraint:onDelete:CASCADE"`
	EvidenceDocs      []EvidenceDocument `json:"evidence_docs" gorm:"foreignKey:ProfileID;constraint:onDelete:CASCADE"`
	EvidenceAIReviews []EvidenceAIReview `json:"evidence_ai_reviews" gorm:"foreignKey:ProfileID;constraint:onDelete:CASCADE"`
}
