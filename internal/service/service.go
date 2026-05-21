package service

import (
	"greentrust-hackathon/internal/blockchain"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/pkg/bcrypt"
	"greentrust-hackathon/pkg/jwt"
	"greentrust-hackathon/pkg/supabase"
	"log"
)

type Service struct {
	UserService          IUserService
	InvestorService      IInvestorService
	EvidenceService      IEvidenceService
	GreenPassportService IGreenPassportService
}

func NewService(repository *repository.Repository, bcrypt bcrypt.Interface, jwtAuth jwt.Interface, supabase supabase.Interface) *Service {
	chainClient, err := blockchain.NewEthereumClientFromEnv()
	if err != nil {
		log.Printf("warning: blockchain client disabled: %v", err)
	}

	userService := NewUserService(repository.UserRepository, repository.OTPRepository, jwtAuth, bcrypt, repository.UserIdentityRepository, repository.BusinessSectorRepository, repository.UMKMProfileRepository, repository.LocationPhotoRepository, supabase)
	investorService := NewInvestorService(repository.UserRepository, repository.UserIdentityRepository, jwtAuth, repository.InvestorProfileRepository, repository.InvestorPositionRepository, repository.SkillRepository)
	evidenceService := NewEvidenceService(repository.UMKMProfileRepository, repository.EvidenceRepository, supabase)
	greenPassportService := NewGreenPassportService(repository.UMKMProfileRepository, repository.EvidenceRepository, repository.GreenPassportRepository, chainClient, supabase)

	return &Service{
		UserService:          userService,
		InvestorService:      investorService,
		EvidenceService:      evidenceService,
		GreenPassportService: greenPassportService,
	}
}
