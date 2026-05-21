package entity

import (
	"time"

	"github.com/google/uuid"
)

type UMKMProduct struct {
	ProductID          uuid.UUID `json:"product_id" gorm:"type:varchar(36);primaryKey"`
	ProfileID          uuid.UUID `json:"profile_id" gorm:"type:varchar(36);not null;index"`
	ProductName        string    `json:"product_name" gorm:"type:varchar(200);not null"`
	ProductDescription string    `json:"product_description" gorm:"type:text;not null"`
	Price              float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	PhotoURL           string    `json:"photo_url" gorm:"type:text;not null"`
	CreatedAt          time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"type:timestamp"`
}
