package repository

import (
	"greentrust-hackathon/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ILocationPhotoRepository interface {
	GetLocationPhotosByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]entity.LocationPhoto, error)
	CreateLocationPhotos(tx *gorm.DB, photos []entity.LocationPhoto) error
	DeleteLocationPhotosByProfileID(tx *gorm.DB, profileID uuid.UUID) error
}

type LocationPhotoRepository struct {
	db *gorm.DB
}

func NewLocationPhotoRepository(db *gorm.DB) ILocationPhotoRepository {
	return &LocationPhotoRepository{db: db}
}

func (r *LocationPhotoRepository) GetLocationPhotosByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]entity.LocationPhoto, error) {
	var photos []entity.LocationPhoto
	err := tx.Debug().Where("profile_id = ?", profileID).Find(&photos).Error
	if err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *LocationPhotoRepository) CreateLocationPhotos(tx *gorm.DB, photos []entity.LocationPhoto) error {
	err := tx.Debug().Create(&photos).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *LocationPhotoRepository) DeleteLocationPhotosByProfileID(tx *gorm.DB, profileID uuid.UUID) error {
	err := tx.Debug().Where("profile_id = ?", profileID).Delete(&entity.LocationPhoto{}).Error
	if err != nil {
		return err
	}
	return nil
}
