package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"

	"gorm.io/gorm"
)

type IUMKMProfileRepository interface {
	GetUMKMProfile(tx *gorm.DB, param model.GetUMKMProfileParam) (*entity.UMKMProfile, error)
	CreateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error
	UpdateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error
}

type UMKMProfileRepository struct {
	db *gorm.DB
}

func NewUMKMProfileRepository(db *gorm.DB) IUMKMProfileRepository {
	return &UMKMProfileRepository{db: db}
}

func (r *UMKMProfileRepository) GetUMKMProfile(tx *gorm.DB, param model.GetUMKMProfileParam) (*entity.UMKMProfile, error) {
	var profile entity.UMKMProfile
	err := tx.Debug().Where(&param).First(&profile).Error
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *UMKMProfileRepository) CreateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error {
	err := tx.Debug().Create(profile).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *UMKMProfileRepository) UpdateUMKMProfile(tx *gorm.DB, profile *entity.UMKMProfile) error {
	err := tx.Debug().Save(profile).Error
	if err != nil {
		return err
	}

	return nil
}
