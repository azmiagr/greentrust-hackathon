package entity

import (
	"time"

	"github.com/google/uuid"
)

type InvestorProfile struct {
	ProfileID    uuid.UUID `json:"profile_id" gorm:"type:varchar(36);primaryKey"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:varchar(36);not null;uniqueIndex"`
	InvestorType string    `json:"investor_type" gorm:"type:varchar(100)"`
	TicketRange  string    `json:"ticket_range" gorm:"type:varchar(100)"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"type:timestamp"`

	Positions    []InvestorPosition `json:"positions" gorm:"foreignKey:ProfileID;constraint:onDelete:CASCADE"`
	FocusSectors []BusinessSector   `json:"focus_sectors" gorm:"many2many:investor_profile_focus_sectors;foreignKey:ProfileID;joinForeignKey:ProfileID;References:SectorID;joinReferences:SectorID"`
}
