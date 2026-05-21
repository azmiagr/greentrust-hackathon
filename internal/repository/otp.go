package repository

import (
	"greentrust-hackathon/entity"
	"greentrust-hackathon/model"

	"gorm.io/gorm"
)

type IOTPRepository interface {
	GetOTP(tx *gorm.DB, param model.GetOtp) (*entity.OTP, error)
	CreateOTP(tx *gorm.DB, otp *entity.OTP) error
	UpdateOTP(tx *gorm.DB, otp *entity.OTP) error
	DeleteOTP(tx *gorm.DB, otp *entity.OTP) error
	DeleteOTPByUserID(tx *gorm.DB, param model.GetOtp) error
}

type OTPRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) IOTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) GetOTP(tx *gorm.DB, param model.GetOtp) (*entity.OTP, error) {
	var otp entity.OTP
	err := tx.Debug().Where(&param).First(&otp).Error
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *OTPRepository) CreateOTP(tx *gorm.DB, otp *entity.OTP) error {
	err := tx.Debug().Create(otp).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *OTPRepository) UpdateOTP(tx *gorm.DB, otp *entity.OTP) error {
	err := tx.Debug().Updates(otp).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *OTPRepository) DeleteOTP(tx *gorm.DB, otp *entity.OTP) error {
	err := tx.Debug().Delete(otp).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *OTPRepository) DeleteOTPByUserID(tx *gorm.DB, param model.GetOtp) error {
	err := tx.Debug().Where("user_id = ?", param.UserID).Delete(&entity.OTP{}).Error
	if err != nil {
		return err
	}

	return nil
}
