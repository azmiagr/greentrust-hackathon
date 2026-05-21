package entity

import (
	"time"

	"github.com/google/uuid"
)

type Skill struct {
	SkillID   uuid.UUID `json:"skill_id" gorm:"type:varchar(36);primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp"`
}
