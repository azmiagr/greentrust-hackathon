package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"

	"gorm.io/gorm"
)

type IInvestorProfileRepository interface {
	GetInvestorProfile(tx *gorm.DB, param model.GetInvestorProfileParam) (*entity.InvestorProfile, error)
	CreateInvestorProfile(tx *gorm.DB, profile *entity.InvestorProfile) error
}

type InvestorProfileRepository struct {
	db *gorm.DB
}

func NewInvestorProfileRepository(db *gorm.DB) IInvestorProfileRepository {
	return &InvestorProfileRepository{db: db}
}

func (r *InvestorProfileRepository) GetInvestorProfile(tx *gorm.DB, param model.GetInvestorProfileParam) (*entity.InvestorProfile, error) {
	var profile entity.InvestorProfile
	err := tx.Debug().Where(&param).First(&profile).Error
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *InvestorProfileRepository) CreateInvestorProfile(tx *gorm.DB, profile *entity.InvestorProfile) error {
	err := tx.Debug().Create(profile).Error
	if err != nil {
		return err
	}

	return nil
}
