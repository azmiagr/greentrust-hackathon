package entity

import "github.com/google/uuid"

type BusinessSector struct {
	SectorID   uuid.UUID `json:"sector_id" gorm:"type:varchar(36);primaryKey"`
	SectorName string    `json:"sector_name" gorm:"type:varchar(100);not null"`

	UMKMProfiles []UMKMProfile `json:"umkm_profiles" gorm:"foreignKey:SectorID;constraint:onDelete:CASCADE"`
}
