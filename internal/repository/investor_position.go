package repository

import (
	"greentrust-hackathon/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IInvestorPositionRepository interface {
	CreateInvestorPosition(tx *gorm.DB, position *entity.InvestorPosition) error
	GetInvestorPositionsByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]*entity.InvestorPosition, error)
	GetInvestorPositionByIDAndProfileID(tx *gorm.DB, positionID uuid.UUID, profileID uuid.UUID) (*entity.InvestorPosition, error)
	UpdateInvestorPosition(tx *gorm.DB, position *entity.InvestorPosition) error
	DeleteInvestorPosition(tx *gorm.DB, position *entity.InvestorPosition) error
}

type InvestorPositionRepository struct {
	db *gorm.DB
}

func NewInvestorPositionRepository(db *gorm.DB) IInvestorPositionRepository {
	return &InvestorPositionRepository{db: db}
}

func (r *InvestorPositionRepository) CreateInvestorPosition(tx *gorm.DB, position *entity.InvestorPosition) error {
	err := tx.Debug().Create(position).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *InvestorPositionRepository) GetInvestorPositionsByProfileID(tx *gorm.DB, profileID uuid.UUID) ([]*entity.InvestorPosition, error) {
	var positions []*entity.InvestorPosition
	err := tx.Debug().
		Preload("Skills").
		Where("profile_id = ?", profileID).
		Order("is_current DESC, start_date DESC").
		Find(&positions).Error
	if err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *InvestorPositionRepository) GetInvestorPositionByIDAndProfileID(tx *gorm.DB, positionID uuid.UUID, profileID uuid.UUID) (*entity.InvestorPosition, error) {
	var position entity.InvestorPosition
	err := tx.Debug().
		Preload("Skills").
		Where("position_id = ? AND profile_id = ?", positionID, profileID).
		First(&position).Error
	if err != nil {
		return nil, err
	}

	return &position, nil
}

func (r *InvestorPositionRepository) UpdateInvestorPosition(tx *gorm.DB, position *entity.InvestorPosition) error {
	err := tx.Debug().Save(position).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *InvestorPositionRepository) DeleteInvestorPosition(tx *gorm.DB, position *entity.InvestorPosition) error {
	err := tx.Debug().Delete(position).Error
	if err != nil {
		return err
	}

	return nil
}
