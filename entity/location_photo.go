package entity

import (
	"time"

	"github.com/google/uuid"
)

type LocationPhoto struct {
	PhotoID   uuid.UUID `json:"photo_id" gorm:"type:varchar(36);primaryKey"`
	ProfileID uuid.UUID `json:"profile_id" gorm:"type:varchar(36);not null;index"`
	URL       string    `json:"url" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
}
