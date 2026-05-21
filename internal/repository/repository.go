package repository

import (
	"gorm.io/gorm"
)

type Repository struct {
	UserRepository             IUserRepository
	OTPRepository              IOTPRepository
	UserIdentityRepository     IUserIdentityRepository
	BusinessSectorRepository   IBusinessSectorRepository
	LocationPhotoRepository    ILocationPhotoRepository
	UMKMProfileRepository      IUMKMProfileRepository
	InvestorProfileRepository  IInvestorProfileRepository
	InvestorPositionRepository IInvestorPositionRepository
	SkillRepository            ISkillRepository
	EvidenceRepository         IEvidenceRepository
	GreenPassportRepository    IGreenPassportRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		UserRepository:             NewUserRepository(db),
		OTPRepository:              NewOTPRepository(db),
		UserIdentityRepository:     NewUserIdentityRepository(db),
		BusinessSectorRepository:   NewBusinessSectorRepository(db),
		LocationPhotoRepository:    NewLocationPhotoRepository(db),
		UMKMProfileRepository:      NewUMKMProfileRepository(db),
		InvestorProfileRepository:  NewInvestorProfileRepository(db),
		InvestorPositionRepository: NewInvestorPositionRepository(db),
		SkillRepository:            NewSkillRepository(db),
		EvidenceRepository:         NewEvidenceRepository(db),
		GreenPassportRepository:    NewGreenPassportRepository(db),
	}
}
