package repository

import (
	"greentrust-hackathon/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IGreenPassportRepository interface {
	GetByProfileID(tx *gorm.DB, profileID uuid.UUID) (*entity.GreenPassport, error)
	UpsertPassport(tx *gorm.DB, passport *entity.GreenPassport) error
}

type GreenPassportRepository struct{}

func NewGreenPassportRepository(db *gorm.DB) IGreenPassportRepository {
	return &GreenPassportRepository{}
}

func (r *GreenPassportRepository) GetByProfileID(tx *gorm.DB, profileID uuid.UUID) (*entity.GreenPassport, error) {
	var passport entity.GreenPassport
	err := tx.Debug().Where("profile_id = ?", profileID).First(&passport).Error
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

func (r *GreenPassportRepository) UpsertPassport(tx *gorm.DB, passport *entity.GreenPassport) error {
	err := tx.Debug().Save(passport).Error
	if err != nil {
		return err
	}
	return nil
}
