package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"

	"gorm.io/gorm"
)

type IBusinessSectorRepository interface {
	GetBusinessSector(tx *gorm.DB, param model.GetBusinessSectorParam) (*entity.BusinessSector, error)
}

type BusinessSectorRepository struct {
	db *gorm.DB
}

func NewBusinessSectorRepository(db *gorm.DB) IBusinessSectorRepository {
	return &BusinessSectorRepository{db: db}
}

func (r *BusinessSectorRepository) GetBusinessSector(tx *gorm.DB, param model.GetBusinessSectorParam) (*entity.BusinessSector, error) {
	var sector entity.BusinessSector
	err := tx.Debug().Where(&param).First(&sector).Error
	if err != nil {
		return nil, err
	}

	return &sector, nil
}
