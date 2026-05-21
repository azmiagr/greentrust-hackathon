package repository

import (
	"gorm.io/gorm"
)

type Repository struct {
	UserRepository IUserRepository
	OTPRepository  IOTPRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository: NewUserRepository(db),
		OTPRepository:  NewOTPRepository(db),
	}
}
