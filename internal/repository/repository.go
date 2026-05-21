package repository

import (
	"gorm.io/gorm"
)

type Repository struct {
	UserRepository           IUserRepository
	OTPRepository            IOTPRepository
	UserIdentityRepository   IUserIdentityRepository
	BusinessSectorRepository IBusinessSectorRepository
	LocationPhotoRepository  ILocationPhotoRepository
	UMKMProfileRepository    IUMKMProfileRepository
	EvidenceRepository       IEvidenceRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:           NewUserRepository(db),
		OTPRepository:            NewOTPRepository(db),
		UserIdentityRepository:   NewUserIdentityRepository(db),
		BusinessSectorRepository: NewBusinessSectorRepository(db),
		LocationPhotoRepository:  NewLocationPhotoRepository(db),
		UMKMProfileRepository:    NewUMKMProfileRepository(db),
		EvidenceRepository:       NewEvidenceRepository(db),
	}
}
