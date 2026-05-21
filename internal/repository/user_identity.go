package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"

	"gorm.io/gorm"
)

type IUserIdentityRepository interface {
	GetUserIdentity(tx *gorm.DB, param model.GetUserIdentityParam) (*entity.UserIdentity, error)
	CreateUserIdentity(tx *gorm.DB, identity *entity.UserIdentity) error
	UpdateUserIdentity(tx *gorm.DB, identity *entity.UserIdentity) error
}

type UserIdentityRepository struct {
	db *gorm.DB
}

func NewUserIdentityRepository(db *gorm.DB) IUserIdentityRepository {
	return &UserIdentityRepository{
		db: db,
	}
}

func (r *UserIdentityRepository) GetUserIdentity(tx *gorm.DB, param model.GetUserIdentityParam) (*entity.UserIdentity, error) {
	var identity entity.UserIdentity
	err := tx.Debug().Where(&param).First(&identity).Error
	if err != nil {
		return nil, err
	}

	return &identity, nil
}

func (r *UserIdentityRepository) CreateUserIdentity(tx *gorm.DB, identity *entity.UserIdentity) error {
	err := tx.Debug().Create(identity).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserIdentityRepository) UpdateUserIdentity(tx *gorm.DB, identity *entity.UserIdentity) error {
	err := tx.Debug().Save(identity).Error
	if err != nil {
		return err
	}

	return nil
}
