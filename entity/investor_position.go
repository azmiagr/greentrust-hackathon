package entity

import (
	"time"

	"github.com/google/uuid"
)

type InvestorPosition struct {
	PositionID      uuid.UUID  `json:"position_id" gorm:"type:varchar(36);primaryKey"`
	ProfileID       uuid.UUID  `json:"profile_id" gorm:"type:varchar(36);not null;index"`
	Title           string     `json:"title" gorm:"type:varchar(150);not null"`
	InstitutionName string     `json:"institution_name" gorm:"type:varchar(200);not null"`
	EmploymentType  string     `json:"employment_type" gorm:"type:varchar(50);not null"`
	Location        string     `json:"location" gorm:"type:varchar(200)"`
	StartDate       time.Time  `json:"start_date" gorm:"type:date;not null"`
	EndDate         *time.Time `json:"end_date" gorm:"type:date"`
	IsCurrent       bool       `json:"is_current" gorm:"type:boolean"`
	Description     string     `json:"description" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"type:timestamp"`

	Skills []Skill `json:"skills" gorm:"many2many:investor_position_skills;foreignKey:PositionID;joinForeignKey:PositionID;References:SkillID;joinReferences:SkillID"`
}
